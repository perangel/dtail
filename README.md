d(e)tail
======

Like `tail`, but with more details!

`dtail` is a cli-tool for realtime monitoring of structured log files (e.g. HTTP access logs).

Features
--------

* Supports follow with retry, similar to `tail -F`
* Configurable alert via Monitors (see: `pkg/monitor`)
  * Notifies when alert is triggered
  * Notifies when alert is resolved
* Prints a simple report of request traffic at a configurable interval

Installation
------------

```
go get github.com/perangel/dtail
```

via docker

```
docker build -t ddtail .
```

Usage
-----

```
Usage:
  dtail [FILE] [flags]

Flags:
  -t, --alert-threshold float         Threshold value for triggering an alert during the monitor's alert window. (default 10)
  -w, --alert-window duration         Time frame for evaluating a metric against the alert threshold. (default 2m0s)
  -h, --help                          help for dtail
  -r, --monitor-resolution duration   Monitor resolution (e.g. 30s, 1m, 5h) (default 1s)
  -i, --report-interval int           Print a report every -i seconds. (default 10)
  -F, --retry-follow tail -F          Retry file after rename or deletion. Similar to tail -F.
  ```

Run via docker

**NOTE:** There is an [issue](https://github.com/docker/for-mac/issues/2375) with filesystem events not triggering on mounted volumes on docker-for-mac. As a result, you'll need to write to the source file (e.g. /tmp/access.log) from inside the container to work around this.

```
// see note above on limitation with mounted volumes
docker run -it -v /tmp:/tmp ddtail -F /tmp/access.log
```

TODO
----

* Performance testing and benchmarks
* Add more test coverage
* Add support for reading from `stdin`
* Refactor core logic in main.go into `pkg/dtail`
* Add support for monitor alert message templates (e.g. on warn, on resolve)
* Add support for multiple monitors
* Add support for simple dsl/query language for configuring monitors via command-line or config file
* Add support for StatsD 
* Add support for configurable parsers (currently only supports Common Log format)
* Refactor reporting logic to support templates
* Improve error handling in a few places (e.g. don't just ignore)
* Workaround for fs events on mounted volumes [issue](https://github.com/docker/for-mac/issues/2375)
