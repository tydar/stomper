# Stomper

![Actions Status](https://github.com/tydar/stomper/actions/workflows/go.yml/badge.svg)

A Go message queue implementing the [STOMP protocol](https://stomp.github.io/stomp-specification-1.2.html).

## Examples

Currently one simple point-to-point messaging example is included at `examples/PointToPoint`. Clone the repo, navigate to that folder, and run `docker-compose up` to see it in action.

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
| TCPDeadline | STOMPER_TCPDEADLINE | 30 | TCP timeout (time in seconds allowed between msg from client, 0 means no timeout) |
| LogPath   | STOMPER_LOGPATH | ./stomper.log | path to log file |
| LogToFile | STOMPER_LOGTOFILE | true     | should we stomper log to a file? |
| LogToStdout| STOMPER_LOGTOSTDOUT| false   | should stomper log to stdout? |
| Topics    | STOMPER_TOPICS    | ["/queue/main"] | list of pub-sub topics |
| SendWorkers| STOMPER_SENDWORKERS| 5 | Number of send worker goroutines to spawn |

An example config file is provided: `stomper_config.yaml`. Stomper will look for a file with this name in either `/etc/stomper` or the directory from which it is called.

## Notes beyond STOMP specification

* Creating a topic with a message
    * If a SUBSCRIBE frame includes a header with key `create` and value `true`, it will create the topic if it does not already exist.
## Done

* Frame parsing
* Define interface for queueing
* Implement memory queue backend
* Frame handling
  * CONNECT
  * SUBSCRIBE
  * UNSUBSCRIBE
  * SEND
  * MESSAGE
  * RECEIPT
  * ERROR
* Semantics
  * Supports only pub-sub currently

## TODO

* Server connection protocol
    * Size limits?
    * Auth?
        * crypto/tls
* Define semantics beyond STOMP protocol
* Implement frame actions
    * BEGIN
    * COMMIT
    * ABORT
    * ACK
    * NACK
    * DISCONNECT
* Message queueing (depends on ACK, NACK)
* Configuration of worker pool for message forwarding
