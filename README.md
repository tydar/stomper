# Stomper

![Actions Status](https://github.com/tydar/stomper/actions/workflows/go.yml/badge.svg)

A Go message queue implementing the [STOMP protocol](https://stomp.github.io/stomp-specification-1.2.html).

## Install and test latest build from GHCR

1. Pull from Docker:

```shell
$ docker pull ghcr.io/tydar/stomper:main
```

2. Run and expose the port from the example config:

```shell
$ docker run -d -p 32801:32801 ghcr.io/tydar/stomper:main
```

3. Connect with a STOMP client at port 32801 and test!

To overwrite Stomper's default settings, you can pass environment variables to the container:

```shell
$ docker run -d -p 32801:32801 --env STOMPER_TOPICS="/queue/env1 /queue/env2" ghcr.io/tydar/stomper:main
```

## Configuration options

| Parameter | ENV variable | Default Value | Description |
| --------- | ------------ | ------------- | ----------- |
| Port      | STOMPER_PORT | 32801         | TCP port server listens on |
| Hostname  | STOMPER_HOSTNAME | localhost | hostname on which server accepts connections |
| TCPDeadline | STOMPER_TCPDEADLINE | 30 | TCP timeout (time between messages from client) |
| LogPath   | STOMPER_LOGPATH | ./stomper.log | path to log file |
| LogToFile | STOMPER_LOGTOFILE | true     | should we stomper log to a file? |
| LogToStdout| STOMPER_LOGTOSTDOUT| false   | should stomper log to stdout? |
| Topics    | STOMPER_TOPICS    | ["/queue/main"] | list of pub-sub topics |

An example config file is provided: `stomper_config.yaml`. Stomper will look for a file with this name in either `/etc/stomper` or the directory from which it is called.

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
