package tail

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cloudflare/cfssl/log"
	"github.com/fsnotify/fsnotify"
)

type Config struct {
	// If true, requires that the file being tailed exist when calling TailFile(), otherwise it will immediately return an error.
	FileMustExist bool

	// ResumeWatching controls whether the Tail should wait to resume watching the file in the event that it is removed or renamed.
	// This would be useful if you wanted to support log-rotation of the target file.
	ResumeWatching bool
}

type Tail struct {
	Config *Config
	Lines  chan string
	Errors chan error

	file    *os.File
	reader  *bufio.Reader
	watcher *fsnotify.Watcher

	shutdownCh chan os.Signal
	doneCh     chan bool
}

// waitForFile will attempt to os.Stat() a file every second until it is present
// or an error other than IsNotExist is returned.
func waitForFile(filepath string) error {
	for {
		_, err := os.Stat(filepath)
		if err != nil {
			if os.IsNotExist(err) {
				time.Sleep(1 * time.Second)
				continue
			} else {
				return err
			}
		}
		break
	}
	return nil
}

func (t *Tail) openFile(filepath string) error {
	f, err := os.OpenFile(filepath, os.O_RDONLY, 0)
	if err != nil {
		if os.IsNotExist(err) {
			if t.Config.FileMustExist {
				return fmt.Errorf("no such file `%s`", filepath)
			}

			err = waitForFile(filepath)
			if err != nil {
				return err
			}
			return t.openFile(filepath)
		}
	}

	t.file = f
	t.reader = bufio.NewReader(f)

	return nil
}

func (t *Tail) tail() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		t.Errors <- fmt.Errorf("failed to initialize a new file watcher: %s", err)
	}
	defer watcher.Close()
	t.watcher = watcher

	go func() {
		for {
			select {
			case event := <-t.watcher.Events:
				if event.Op&fsnotify.Write == fsnotify.Write {
					b, err := t.reader.ReadBytes('\n')
					if err != nil {
						if err == io.EOF {
							t.Errors <- err
						}
					}
					t.Lines <- string(b)
					continue
				}

				if event.Op&fsnotify.Remove == fsnotify.Remove {
					if t.Config.ResumeWatching {
						err := waitForFile(t.file.Name())
						if err != nil {
							t.Errors <- err
						}
						t.watcher.Add(t.file.Name())
					} else {
						t.Errors <- fmt.Errorf("file missing")
						continue
					}
				}

				if event.Op&fsnotify.Rename == fsnotify.Rename {
					// stop watching the current file
					t.watcher.Remove(t.file.Name())
					if !t.Config.FileMustExist {
						if err := waitForFile(t.file.Name()); err != nil {
							t.Errors <- err
						}
						t.watcher.Add(t.file.Name())
						continue
					}
					t.Errors <- fmt.Errorf("file %s has been renamed", t.file.Name())
				}

			case err := <-t.watcher.Errors:
				// TODO: should these be fatal?
				t.Errors <- err

			case err := <-t.Errors:
				// TODO: Configure a logger
				log.Error(err)

			}
		}
	}()

	err = t.watcher.Add(t.file.Name())
	if err != nil {
		t.Errors <- err
	}

	<-t.doneCh
}

func (t *Tail) cleanup() error {
	close(t.Lines)
	close(t.Errors)
	close(t.shutdownCh)
	return t.watcher.Remove(t.file.Name())
}

// TailFile configures a new Tail to follow the specified file.
func TailFile(filepath string, config *Config) (*Tail, error) {
	t := &Tail{
		Config:     config,
		Lines:      make(chan string),
		Errors:     make(chan error),
		shutdownCh: make(chan os.Signal, 1),
	}

	signal.Notify(t.shutdownCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-t.shutdownCh
		t.cleanup()
		t.doneCh <- true
	}()

	err := t.openFile(filepath)
	if err != nil {
		return nil, err
	}

	go t.tail()

	return t, nil
}
