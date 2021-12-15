package main

import (
	"bufio"
	"bytes"
	"io"
	"net"
	"reflect"
	"strings"
	"testing"
)

func TestConnectionManager(t *testing.T) {
	messages := make(chan CnxMgrMsg)
	cm := NewConnectionManager("", 32801, messages, 1)
	err := cm.Start()
	if err != nil {
		t.Error("error starting cnx manager", err)
	}

	t.Run("_ConnectAndRead", func(t *testing.T) {
		conn, err := net.Dial("tcp", ":32801")
		if err != nil {
			t.Error("could not connect to server: ", err)
		}
		message := <-messages

		if message.Type != NEW_CONNECTION {
			t.Error("ConnectionManager did not send message")
		}
		conn.Close()
		message = <-messages
	})

	t.Run("_Write", func(t *testing.T) {
		conn, err := net.Dial("tcp", ":32801")
		if err != nil {
			t.Error("could not connect to server: ", err)
		}
		msg := <-messages

		if msg.Type != NEW_CONNECTION {
			t.Errorf("Did not receive new connection message from ConnectionManager: %d %s", msg.Type, msg.Msg)
		}

		id := msg.Msg
		err = cm.Write(id, "Test\n")
		if err != nil {
			t.Error("write error: ", err)
		}

		b := bytes.NewBuffer([]byte{})
		_, err = io.CopyN(b, conn, 5)
		if err != nil {
			t.Error("read error: ", err)
		}

		if b.String() != "Test\n" {
			t.Errorf("got %s wanted Test\n", b.String())
		}
		conn.Close()
		msg = <-messages
	})

	t.Run("_handleRemovals", func(t *testing.T) {
		conn, err := net.Dial("tcp", ":32801")
		if err != nil {
			t.Error("could not connect to server: ", err)
		}
		msg := <-messages
		conn.Close()
		msg = <-messages
		if msg.Type != CONNECTION_CLOSED {
			t.Error("connection failed to close or another message sent")
		}
	})

	t.Run("_Disconnect", func(t *testing.T) {
		conn, err := net.Dial("tcp", ":32801")
		if err != nil {
			t.Error("coult not connect to server: ", err)
		}
		defer conn.Close()

		msg := <-messages
		err = cm.Disconnect(msg.ID)
		msg = <-messages

		if err != nil {
			t.Error("could not disconnect from client: ", err)
		}

		if msg.Type != CONNECTION_CLOSED {
			t.Error("connection failed to close or another message sent")
		}
	})
}

func TestScanNullTerm(t *testing.T) {
	empty := []byte("")
	nullTerm := []byte("Null-term\000")
	multipleNull := []byte("Null-term\000another\000")
	noNull := []byte("Justwords\n")
	var tests = []struct {
		input []byte
		eof   bool
		adv   int
		token []byte
		err   error
	}{
		{empty, false, 0, nil, nil},
		{nullTerm, false, 10, nullTerm[:len(nullTerm)-1], nil},
		{multipleNull, false, 10, nullTerm[:len(nullTerm)-1], nil},
		{noNull, false, 0, nil, nil},
		{noNull, true, 10, noNull, nil},
	}

	for _, tt := range tests {
		testname := string(tt.input)
		t.Run(testname, func(t *testing.T) {
			advance, token, err := ScanNullTerm(tt.input, tt.eof)
			if err != tt.err {
				t.Errorf("err val: got %v wanted %v\n", err, tt.err)
			} else if advance != tt.adv {
				t.Errorf("wrong adv val: got %d wanted %d\n", advance, tt.adv)
			} else if !reflect.DeepEqual(token, tt.token) {
				t.Errorf("wrong token val: got %s wanted %s\n", string(token), string(tt.token))
			}
		})
	}
}

func TestScannerNullTerm(t *testing.T) {
	nullTerm := "Null-term\000"
	multipleNull := "Null-term\000another\000"
	var tests = []struct {
		input  string
		tokens []string
	}{
		{nullTerm, []string{"Null-term"}},
		{multipleNull, []string{"Null-term", "another"}},
	}

	for _, tt := range tests {
		testname := string(tt.input)
		t.Run(testname, func(t *testing.T) {
			scanner := bufio.NewScanner(strings.NewReader(tt.input))
			scanner.Split(ScanNullTerm)
			output := make([]string, 0)
			for scanner.Scan() {
				output = append(output, scanner.Text())
			}
			if !reflect.DeepEqual(output, tt.tokens) {
				t.Errorf("output: %v wanted %v\n", output, tt.tokens)
			}
		})
	}
}
