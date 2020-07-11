# Tripper

Tripper is an HTTP analyser which could be used 
to show the HTTP performance.

It's a study project on Golang and take inspiration from 
https://github.com/moderation/httpstat

More than this it is trying to test some other topics:

- goroutines
- flag parameter
- JSON Marshaling
- input from cli
- log options

```bash

Usage: ./tripper [OPTIONS] URL

OPTIONS:
  -debug
        set debug to true, and print out logs
  -json
        if set, output a json result
  -url string
        a url (default "https://www.google.com")
```

The *debug* options print out all the log calls,
which by default are disabled.

By default it try to query httpp://www.google.com
and print the results:

```bash
Check connection data for :      https://www.google.com

DNS lookup       1.007537ms
TCP connection   58.755528ms
TLS handshake    84.716681ms
ttfb             235.849912ms

Tripper took    2.292134563s
```

there's also a flag to output a JSON object with the same informations.

```JSON
{
    "dnsLookUp": "942.023µs",
    "tcpConnect": "100.798718ms",
    "tlsHandshake": "119.532558ms",
    "ttfb": "344.141836ms",
    "took": "2.273526306s"
}
```