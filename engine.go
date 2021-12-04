package main

import (
	"log"
)

type Engine struct {
	CM       *ConnectionManager
    SM       *SubscriptionManager
	Incoming chan CnxMgrMsg
	Store    *Store
}

func NewEngine(st Store, cm *ConnectionManager, inc chan CnxMgrMsg) *Engine {
	return &Engine{
		CM:       cm,
		Store:    &st,
		Incoming: inc,
        SM:       NewSubscriptionManager(),
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
            case SUBSCRIBE:
                e.handleSubscribe(msg, frame)
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

func (e *Engine) handleSubscribe(msg CnxMgrMsg, frame Frame) {
    clientID := msg.ID
    subID, prs := frame.Headers["id"]
    if !prs {
        log.Printf("Error: no id header on susbcribe message")
    }
    dest, prs := frame.Headers["destination"]
    if !prs {
        log.Printf("Error: no id header on susbcribe message")
    }
    // TODO: add destination validation
    e.SM.Subscribe(clientID, subID, dest) 
}
