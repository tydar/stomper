package main

import (
	"fmt"
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
				log.Printf("ERROR: client %s and error %s\n", msg.ID, err)
				err2 := e.handleError(msg, err)
				if err2 != nil {
					log.Printf("ERROR: client %s write error: %s\n", msg.ID, err2)
				}
			}
			switch frame.Command {
			case CONNECT:
			case STOMP:
				response := e.handleConnect(msg)
				err = e.CM.Write(msg.ID, response)
				if err != nil {
					log.Printf("ERROR: client %s write error: %s\n", msg.ID, err)
				}
			case SUBSCRIBE:
				err = e.handleSubscribe(msg, frame)
				if err != nil {
					log.Println(err)
					err2 := e.handleError(msg, err)
					if err2 != nil {
						log.Printf("ERROR: client %s write error: %s\n", msg.ID, err2)
					}
				}
			case UNSUBSCRIBE:
				err = e.handleUnsubscribe(msg, frame)
				if err != nil {
					log.Println(err)
					err2 := e.handleError(msg, err)
					if err2 != nil {
						log.Printf("ERROR: client %s write error: %s\n", msg.ID, err2)
					}
				}
            case SEND:
                // 2 steps
                // synchronous: process SEND frame and create MESSAGE frame
                // asynchronous: perform the send
                // TODO: create a way for send logic to be decoupled
                err = e.handleSend(msg, frame)
                if err != nil {
                    log.Println(err)
                    err2 := e.handleError(msg, err)
                    if err2 != nil {
                        log.Printf("ERROR: client %s write error: %s\n", msg.ID, err2)
                    }
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
		Headers: map[string]string{"accept-version": "1.2"},
		Body:    "",
	})
}

func (e *Engine) handleSubscribe(msg CnxMgrMsg, frame Frame) error {
	clientID := msg.ID
	subID, prs := frame.Headers["id"]
	if !prs {
		return fmt.Errorf("Error: client %s: no ID on SUBSCRIBE frame", msg.ID)
	}
	dest, prs := frame.Headers["destination"]
	if !prs {
		return fmt.Errorf("Error: client %s: no destination on SUBSCRIBE frame", msg.ID)
	}
	// TODO: add destination validation
	return e.SM.Subscribe(clientID, subID, dest)
}

func (e *Engine) handleUnsubscribe(msg CnxMgrMsg, frame Frame) error {
	clientID := msg.ID
	subID, prs := frame.Headers["id"]

	if !prs {
		return fmt.Errorf("Error: client %s: no ID on UNSUBSCRIBE frame", msg.ID)
	}

	return e.SM.Unsubscribe(clientID, subID)
}

func (e *Engine) handleError(msg CnxMgrMsg, err error) error {
	eFrame := UnmarshalFrame(Frame{
		Command: ERROR,
		Headers: map[string]string{"message": err.Error()},
		Body:    "Original frame: " + msg.Msg,
	})
	return e.CM.Write(msg.ID, eFrame)
}

func (e *Engine) handleSend(msg CnxMgrMsg, frame Frame) error {
    newHeaders := make(map[string]string)
    for k,v := range frame.Headers {
        newHeaders[k] = v
    }

    dest, prs := frame.Headers["destination"]
    if !prs {
        return fmt.Errorf("ERROR: client %s: no destination header", msg.ID)
    }

    messageFrame := UnmarshalFrame(Frame{
        Command: MESSAGE,
        Headers: newHeaders,
        Body: frame.Body,
    })


    // start send worker
    // simple pub-sub architecture here
    // TODO: limit on concurrency?
    subscribers := e.SM.ClientsByDestination(dest)
    log.Printf("SENDING_MESSAGE: originator ID %s to %d subscribers\n", msg.ID, len(subscribers))
    
    go func(subscribers []string, fr string) {
        for i := range subscribers {
            errWrite := e.CM.Write(subscribers[i], messageFrame)
            if errWrite != nil {
                log.Printf("ERROR: client %s: MESSAGE send failed\n", subscribers[i])
            }
        }
    }(subscribers, messageFrame)

    return nil
}
