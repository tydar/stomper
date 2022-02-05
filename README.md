# Stomper

![Actions Status](https://github.com/tydar/stomper/actions/workflows/go.yml/badge.svg)

A message broker implementing the [STOMP protocol](https://stomp.github.io/stomp-specification-1.2.html) written in Go.

## Examples

I have written a simple chat client designed to use a generic STOMP pub-sub server called [stomp-chat](https://github.com/tydar/stomp-chat) that can serve as an example of a client for this server.

Currently one simple point-to-point messaging example is included at `examples/PointToPoint`. Clone the repo, navigate to that folder, and run `docker-compose up` to see it in action.

## Install and test latest build from GHCR

1. Pull Docker image from the GHCR:

```shell
$ docker pull ghcr.io/tydar/stomper:main
```

2. Run and publish the port from the example config:

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
| TCPDeadline | STOMPER_TCPDEADLINE | 0 | TCP timeout (time in seconds allowed between msg from client, 0 means no timeout) |
| LogPath   | STOMPER_LOGPATH | ./stomper.log | path to log file |
| LogToFile | STOMPER_LOGTOFILE | true     | should we stomper log to a file? |
| LogToStdout| STOMPER_LOGTOSTDOUT| false   | should stomper log to stdout? |
| Topics    | STOMPER_TOPICS    | ["/queue/main"] | list of pub-sub topics |
| SendWorkers| STOMPER_SENDWORKERS| 1 | Number of send worker goroutines to spawn |
| MetricsServer| STOMPER_METRICSERVER| false | should we expose a JSON metrics endpoint? |
| MetricsAddress | STOMPER_METRICSADDRESS | ":8080" | address string for Metrics Service ListenAndServe call|

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
  * DISCONNECT
  * BEGIN
  * COMMIT
  * ABORt
* Semantics
  * Supports only pub-sub currently
* runtime topic creation by clients
* Configuration of worker pool for message forwarding


## TODO

* Server connection protocol
    * Size limits?
    * Rate limits?
    * Auth?
        * crypto/tls
* RBAC?
* Define semantics beyond STOMP protocol
* Implement frame actions
    * ACK
    * NACK
* Message queueing (depends on ACK, NACK)



