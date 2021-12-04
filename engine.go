package main

import (
	"log"
)

type Engine struct {
	CM       *ConnectionManager
	Clients  []string // may want to spin this off to its own Client type or ClientManager interface
	Incoming chan CnxMgrMsg
	Store    *Store
}

func NewEngine(st Store, cm *ConnectionManager, inc chan CnxMgrMsg) *Engine {
	return &Engine{
		CM:       cm,
		Store:    &st,
		Incoming: inc,
		Clients:  make([]string, 0),
	}
}

func (e *Engine) Start() error {
	err := e.CM.Start()
	if err != nil {
		return err
	}

	log.Println("Entering main loop")
	for msg := range e.Incoming {
		if msg.Type == FRAME {
			frame, err := ParseFrame(msg.Msg)
			if err != nil {
				// for now, just stop if we get an error
				// need to handle sending ERROR frames and moving on later
				log.Fatal(err)
			}
			switch frame.Command {
			case CONNECT:
			case STOMP:
				response := e.handleConnect(msg)
				err = e.CM.Write(msg.ID, response)
				if err != nil {
					log.Fatal(err)
					//see above
				}
			}
		}
	}

	return nil
}

func (e *Engine) handleConnect(msg CnxMgrMsg) string {
	// e.handleConnect takes a CONNECT or STOMP frame and produces a CONNECTED frame
	return UnmarshalFrame(Frame{
		Command: CONNECTED,
		Headers: map[string]string{"version": "1.2"},
		Body:    "",
	})
}
