d(e)tail
======

Like `tail -F`, but with more details!

`dtail` is a cli utility for realtime monitoring of structured log files (e.g. HTTP access logs).

Features
--------

* Follow with retry, a la `tail -F`
* Alerting via Monitors 



* Computes summary statistics about the traffic:
    * TopN most requested pages by "section" (e.g. for path /pages/create, section is /pages)
    * TopN most active remote clients
    * TopN most active users
    * Error rate (4xx and 5xx status codes)

* Notify with an alert if average requests per second (RPS) exceeds a given threshold within a specified interval (e.g. Avg. requests per second exceeds 10rps over the past 2 minutes)
    * Notify when average requests per second drops below the threshold


Installation
------------
