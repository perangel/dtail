d(e)tail
======

Like `tail`, but with more details!

`dtail` is a cli-tool for realtime monitoring of structured log files (e.g. HTTP access logs).

Features
--------

* Supports follow with retry, similar to `tail -F`

* Configurable alerts with Monitors (see: `pkg/monitor`)

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
```
docker run -it -v /tmp:/tmp ddtail -F /tmp/access.log
```

TODO
----

* Refactor core logic in main.go into `pkg/dtail`
* Add support for multiple monitors
* Add support for simple dsl/query language for configuring monitors via command-line or config file
* Add more test coverage
* Performance testing and benchmarks
* Add support for statsd
* Support for configurable parsers (e.g. more than just the default Common Log format)
