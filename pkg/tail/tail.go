package tail

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
)

// ErrFileRemoved is an error that will be returned when the tailed file is removed or renamed
var ErrFileRemoved = errors.New("target file no longer exists")

type Config struct {
	// If true, Tail will keep retrying to open a file after it has been renamed or removed.
	// This option is useful when you need to handle logoration.
	Retry bool
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
			if !t.Config.Retry {
				return fmt.Errorf("%v: No such file or directory", filepath)
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
		t.Errors <- fmt.Errorf("fsnotify: %s", err)
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

				if event.Op&fsnotify.Remove == fsnotify.Remove || event.Op&fsnotify.Rename == fsnotify.Rename {
					// stop watching the current file
					t.watcher.Remove(t.file.Name())

					// if retry enabled, wait for the file
					if t.Config.Retry {
						err := waitForFile(t.file.Name())
						if err != nil {
							t.Errors <- err
						}
						t.watcher.Add(t.file.Name())
					} else {
						t.Errors <- ErrFileRemoved
						continue
					}
				}

			case err := <-t.watcher.Errors:
				t.Errors <- err
			}
		}
	}()

	err = t.watcher.Add(t.file.Name())
	if err != nil {
		t.Errors <- err
	}

	<-t.doneCh
}

func (t *Tail) stop() error {
	err := t.watcher.Remove(t.file.Name())
	close(t.Lines)
	close(t.Errors)
	close(t.shutdownCh)
	return err
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
		t.stop()
		t.doneCh <- true
	}()

	err := t.openFile(filepath)
	if err != nil {
		return nil, err
	}

	go t.tail()

	return t, nil
}

// Wait blocks waiting for any errors returned by Tail
func (t *Tail) Wait() error {
	for {
		select {
		case err := <-t.Errors:
			return err
		}
	}
}
