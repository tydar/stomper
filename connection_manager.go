package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"log"
	"net"
	"strconv"
	"sync"

	"github.com/google/uuid"
)

// design notes:
// * want to ensure TCP connections stay decoupled from the main protocol engine
// * so only connections are managed here -- subscriptions, parsing, etc are elsewhere
// * need one goroutine to own each connection.
// * potential type answer: connections map[uuid.UUID]Connection

// ConnectionManager
type ConnectionManager struct {
    listener    net.Listener
	hostname    string
	port        int
	connections map[uuid.UUID]*Connection
	messages    chan CnxMgrMsg
	mu          sync.RWMutex
}

func NewConnectionManager(hostname string, port int, messages chan CnxMgrMsg) *ConnectionManager {
	return &ConnectionManager{
        listener:    nil,
		hostname:    hostname,
		port:        port,
		connections: make(map[uuid.UUID]*Connection),
		messages:    messages,
	}
}

func (cm *ConnectionManager) Start() error {
	l, err := net.Listen("tcp", cm.hostname+":"+strconv.Itoa(cm.port))
	if err != nil {
		return err
	}
    cm.mu.Lock()
    cm.listener = l
    cm.mu.Unlock()

	// since the main loop will wait on new connections, we need to start a goroutine
	// to handle any connections that notify us of a deletion request
	removeConnectionChan := make(chan uuid.UUID)
	go cm.handleRemovals(removeConnectionChan)

	// this avoids tests being blocked
	// not sure if it creates any problems for the actual software
	go func() {
		for {
			conn, err := l.Accept() // loop will wait here until a new connection
			if err != nil {
				log.Fatal(err)
			}

			thisUUID := uuid.New()
			cm.mu.Lock()
			cm.connections[thisUUID] = NewConnection(conn, thisUUID)
			go cm.connections[thisUUID].Read(cm.messages, removeConnectionChan)
			cm.mu.Unlock()

			cm.messages <- CnxMgrMsg{
				Type: NEW_CONNECTION,
				Msg:  thisUUID.String(),
			}
		}
	}()
	return nil
}

func (cm *ConnectionManager) Stop() error {
    return cm.listener.Close()
}

func (cm *ConnectionManager) Write(id uuid.UUID, msg string) error {
	cm.mu.RLock()
	connection, prs := cm.connections[id]
	cm.mu.RUnlock()

	if !prs {
		return errors.New(fmt.Sprintf("Connection %v no longer open.\n", id))
	}

	return connection.Write(msg)
}

func (cm *ConnectionManager) handleRemovals(requests chan uuid.UUID) {
	for id := range requests {
		cm.mu.Lock()
		delete(cm.connections, id)
		cm.mu.Unlock()
        cm.messages <- CnxMgrMsg{
            Type: CONNECTION_CLOSED,
            Msg: id.String(),
        }
	}
}

// Connection
type Connection struct {
	id   uuid.UUID
	conn net.Conn
}

func NewConnection(conn net.Conn, id uuid.UUID) *Connection {
	return &Connection{
		conn: conn,
		id:   id,
	}
}

func (c *Connection) Read(readTo chan CnxMgrMsg, done chan uuid.UUID) {
	scanner := bufio.NewScanner(c.conn)
	scanner.Split(ScanNullTerm)
	for {
		if ok := scanner.Scan(); !ok {
			break
		}
		readTo <- CnxMgrMsg{Type: FRAME,
			Msg: (scanner.Text() + "\000"), // have to append the null byte that the scanner strips
		}
	}
	done <- c.id
}

func (c *Connection) Write(msg string) error {
	_, err := c.conn.Write([]byte(msg))
	return err
}

// CnxMgrMsg
// stub struct for messages sent by the ConnectionManager
const (
	NEW_CONNECTION = iota
	CONNECTION_CLOSED
	FRAME
)

type CnxMgrMsg struct {
	Type int
	Msg  string
}

// Custom scanner to split incoming stream on \000
func ScanNullTerm(data []byte, atEOF bool) (int, []byte, error) {
	// if we're at EOF, we're done for now
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}

	// if we find a '\000', return the data up to and including that index
	if i := bytes.IndexByte(data, '\000'); i >= 0 {
		// there is a null-terminated frame
		return i + 1, data[0:i], nil
	}

	// if we are at EOF and we have data, return it so we can see what's going on
	if atEOF {
		return len(data), data, nil
	}
	return 0, nil, nil
}
