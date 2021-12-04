package main

import (
	"github.com/google/uuid"
	"log"
)

type Engine struct {
	CM       *ConnectionManager
	Clients  []uuid.UUID // may want to spin this off to its own Client type or ClientManager interface
	Incoming chan CnxMgrMsg
	Store    *Store
}

func NewEngine(st Store, cm *ConnectionManager, inc chan CnxMgrMsg) *Engine {
	return &Engine{
		CM:       cm,
		Store:    &st,
		Incoming: inc,
		Clients:  make([]uuid.UUID, 0),
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
				response := UnmarshalFrame(Frame{
					Command: CONNECTED,
					Headers: map[string]string{"version": "1.2"},
					Body:    "",
				})
				err = e.CM.Write(msg.ID, response)
				if err != nil {
					log.Fatal(err)
					//see above
				}
				e.Clients = append(e.Clients, msg.ID)
			}
		}
	}

	return nil
}
