# Stomper

![Actions Status](https://github.com/tydar/stomper/actions/workflows/go.yml/badge.svg)

A Go message queue implementing the [STOMP protocol](https://stomp.github.io/stomp-specification-1.2.html).

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

### Connection Manager

The *connection manager* listens for incoming TCP messages, maintains a list of active connections, and handles keeping open
or closing connections as needed. It also contains the logic for when to send data to the engine by scanning for null terminated
frames. It sends those to the engine for parsing.
