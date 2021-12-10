# Stomper

![Actions Status](https://github.com/tydar/stomper/actions/workflows/go.yml/badge.svg)

A Go message queue implementing the [STOMP protocol](https://stomp.github.io/stomp-specification-1.2.html).

## Install and test latest build from main

1. Pull from Docker:

```shell
$ docker pull ghcr.io/tydar/stomper:main
```

2. Run and expose the port from the example config:

```shell
$ docker run -d -p 32801:32801 ghcr.io/tydar/stomper:main
```

3. Connect with a STOMP client at port 32801 and test!

## Done

* Frame parsing
* Define interface for queueing
* Implement memory queue backend

## TODO

* Server connection protocol
    * Manage connections
    * TTL protocol
    * Size limits?
    * Auth?
        * crypto/tls
* Define semantics beyond STOMP protocol
* Implement frame actions
    * SEND
    * SUBSCRIBE
    * UNSUBSCRIBE
    * BEGIN
    * COMMIT
    * ABORT
    * ACK
    * NACK
    * DISCONNECT
    * MESSAGE
    * RECEIPT
    * ERROR
* Define server initialization
    * When are topics defined?
    * What needs to be configured to enable connections?
* Rewrite parsing information. Does not conform to this direction from the standard:
    * If a content-length header is included, this number of octets MUST be read, regardless of whether or not there are NULL octets in the body.
    * This is because the scanner used by ConnectionManager to tokenize input into frames will stop reading at the first null octet.
