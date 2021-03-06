package main

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"
)

// design notes:
// * want to ensure TCP connections stay decoupled from the main protocol engine
// * so only connections are managed here -- subscriptions, parsing, etc are elsewhere
// * need one goroutine to own each connection.
// * potential type answer: connections map[string]Connection

// ConnectionManager
type ConnectionManager struct {
	listener    net.Listener
	hostname    string
	port        int
	connections map[string]*Connection
	messages    chan CnxMgrMsg
	timeout     time.Duration
	mu          sync.RWMutex
}

func NewConnectionManager(hostname string, port int, messages chan CnxMgrMsg, timeout time.Duration) *ConnectionManager {
	log.Printf("NEW_CONNECTION_MANAGER on %s:%d with timeout %d seconds\n", hostname, port, timeout)
	return &ConnectionManager{
		listener:    nil,
		hostname:    hostname,
		port:        port,
		connections: make(map[string]*Connection),
		messages:    messages,
		timeout:     timeout * time.Second,
	}
}

func (cm *ConnectionManager) Hostname() string {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.hostname
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
	removeConnectionChan := make(chan string)
	go cm.handleRemovals(removeConnectionChan)

	// this avoids tests being blocked
	// not sure if it creates any problems for the actual software
	go func() {
		for {
			conn, err := l.Accept() // loop will wait here until a new connection
			if err != nil {
				log.Fatal(err)
			}

			thisUUID := uuid.NewString()
			cm.mu.Lock()
			cm.connections[thisUUID] = NewConnection(conn, thisUUID)
			go cm.connections[thisUUID].Read(cm.messages, removeConnectionChan, cm.timeout)
			cm.mu.Unlock()

			log.Printf("NEW_CONNECTION: ID %s from remote address %s\n", thisUUID, conn.RemoteAddr().String())

			cm.messages <- CnxMgrMsg{
				Type: NEW_CONNECTION,
				ID:   thisUUID,
				Msg:  thisUUID,
			}
		}
	}()
	return nil
}

func (cm *ConnectionManager) Stop() error {
	return cm.listener.Close()
}

func (cm *ConnectionManager) Write(id string, msg string) error {
	cm.mu.RLock()
	connection, prs := cm.connections[id]
	cm.mu.RUnlock()

	if !prs {
		return fmt.Errorf("Connection %v no longer open", id)
	}

	return connection.Write(msg)
}

func (cm *ConnectionManager) handleRemovals(requests chan string) {
	for id := range requests {
		cm.mu.Lock()
		delete(cm.connections, id)
		cm.mu.Unlock()

		log.Printf("CONNECTION_CLOSED by CM: ID %s\n", id)

		cm.messages <- CnxMgrMsg{
			Type: CONNECTION_CLOSED,
			ID:   id,
			Msg:  id,
		}
	}
}

func (cm *ConnectionManager) Disconnect(id string) error {
	cm.mu.RLock()
	connection, prs := cm.connections[id]
	to := cm.timeout
	cm.mu.RUnlock()

	if prs {
		connection.Disconnect(to)
	} else {
		return fmt.Errorf("no such connection: %s", id)
	}
	return nil
}

// Connection
type Connection struct {
	id   string
	conn net.Conn
}

func NewConnection(conn net.Conn, id string) *Connection {
	return &Connection{
		conn: conn,
		id:   id,
	}
}

func (c *Connection) Read(readTo chan CnxMgrMsg, done chan string, timeout time.Duration) {
	scanner := bufio.NewScanner(c.conn)
	scanner.Split(ScanNullTerm)
	for {
		if ok := scanner.Scan(); !ok {
			break
		}
		txt := scanner.Text()
		if txt != "\n" {
			// if it's not just a heartbeat EOL, send the frame
			readTo <- CnxMgrMsg{
				Type: FRAME,
				ID:   c.id,
				Msg:  (txt + "\000"), // have to append the null byte that the scanner strips
			}
		}
		if timeout > 0 {
			c.conn.SetReadDeadline(time.Now().Add(timeout).Add(500 * time.Millisecond))
		}
	}
	done <- c.id
}

func (c *Connection) Write(msg string) error {
	_, err := c.conn.Write([]byte(msg))
	return err
}

func (c *Connection) Disconnect(timeout time.Duration) error {
	// if a connection asks to disconnect, wait timeout seconds and close the connection
	log.Printf("DISCONNECT from client ID %s\n", c.id)
	time.Sleep(timeout)
	return c.conn.Close()
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
	ID   string
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

	// if we did not find a null-terminated frame, still need to check for \n
	if len(data) > 0 && data[0] == '\n' {
		return 1, []byte{data[0]}, nil
	}

	// if we are at EOF and we have data, return it so we can see what's going on
	if atEOF {
		return len(data), data, nil
	}
	return 0, nil, nil
}
