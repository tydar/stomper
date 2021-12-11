package main

import (
	"fmt"
	"log"
	"strconv"
)

type Engine struct {
	CM       *ConnectionManager
	SM       *SubscriptionManager
	Incoming chan CnxMgrMsg
	Store    Store
}

func NewEngine(st Store, cm *ConnectionManager, inc chan CnxMgrMsg) *Engine {
	return &Engine{
		CM:       cm,
		Store:    st,
		Incoming: inc,
		SM:       NewSubscriptionManager(),
	}
}

func (e *Engine) Start() error {
	err := e.CM.Start()
	if err != nil {
		return err
	}

	// start send worker
	// simple pub-sub architecture here
	// TODO: spin this off into its own fuction
	//       && handle additional worker goroutines configurably

	go func() {
		for {
			dests := e.Store.Destinations()
			for j := range dests {
				dest := dests[j]
				count, err := e.Store.Len(dest)
				if err != nil {
					log.Printf("SEND_ERROR: No such destination\n")
				}
				if count > 0 {
					subscribers := e.SM.ClientsByDestination(dest)
					messageFrame, err := e.Store.Pop(dest)
					if err != nil {
						log.Println(err)
					} else {
						log.Printf("SENDING_MESSAGE: on queue %s to %d subscribers\n", dest, len(subscribers))
						for i := range subscribers {
							sub := subscribers[i]
							uniqueHeaders := make(map[string]string)
							for k, v := range messageFrame.Headers {
								uniqueHeaders[k] = v
							}
							uniqueFrame := Frame{
								Command: messageFrame.Command,
								Headers: uniqueHeaders,
								Body:    messageFrame.Body,
							}
							uniqueFrame.Headers["subscription"] = sub.ID
							messageString := UnmarshalFrame(uniqueFrame)
							errWrite := e.CM.Write(sub.ClientID, messageString)
							if errWrite != nil {
								log.Printf("ERROR: client %s: MESSAGE send failed\n", subscribers[i])
							}
						}
					}
				}
			}
		}
	}()

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
				} else {
					err = e.handleReceipt(msg, frame)
					if err != nil {
						log.Printf("ERROR: client %s write error: %s\n", msg.ID, err)
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
				} else {
					err = e.handleReceipt(msg, frame)
					if err != nil {
						log.Printf("ERROR: client %s write error: %s\n", msg.ID, err)
					}
				}
			case SEND:
				err = e.handleSend(msg, frame)
				if err != nil {
					log.Println(err)
					err2 := e.handleError(msg, err)
					if err2 != nil {
						log.Printf("ERROR: client %s write error: %s\n", msg.ID, err2)
					}
				} else {
					err = e.handleReceipt(msg, frame)
					if err != nil {
						log.Printf("ERROR: client %s write error: %s\n", msg.ID, err)
					}
				}
			case DISCONNECT:
			}
		}
	}
	return nil
}

func (e *Engine) handleConnect(msg CnxMgrMsg) string {
	// e.handleConnect takes a CONNECT or STOMP frame and produces a CONNECTED frame
	// TODO: handle protocol negotiation ERROR generation
	heartbeatStr := "0"
	if e.CM.timeout.Milliseconds() > 0 {
		heartbeatStr = strconv.Itoa(int(e.CM.timeout.Milliseconds()))
	} 
	return UnmarshalFrame(Frame{
		Command: CONNECTED,
		Headers: map[string]string{"version": "1.2", "host": e.CM.Hostname(), "heart-beat": "0," + heartbeatStr},
		Body:    "",
	})
}

func (e *Engine) handleSubscribe(msg CnxMgrMsg, frame Frame) error {
	clientID := msg.ID
	subID, prs := frame.Headers["id"]
	if !prs {
		return fmt.Errorf("error: client %s: no ID on SUBSCRIBE frame", msg.ID)
	}
	dest, prs := frame.Headers["destination"]
	if !prs {
		return fmt.Errorf("error: client %s: no destination on SUBSCRIBE frame", msg.ID)
	}
	// TODO: add destination validation
	return e.SM.Subscribe(clientID, subID, dest)
}

func (e *Engine) handleUnsubscribe(msg CnxMgrMsg, frame Frame) error {
	clientID := msg.ID
	subID, prs := frame.Headers["id"]

	if !prs {
		return fmt.Errorf("error: client %s: no ID on UNSUBSCRIBE frame", msg.ID)
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

func (e *Engine) handleReceipt(msg CnxMgrMsg, frame Frame) error {
	receiptID, prs := frame.Headers["receipt"]
	if prs {
		rFrame := UnmarshalFrame(Frame{
			Command: RECEIPT,
			Headers: map[string]string{"receipt-id": receiptID},
			Body:    "",
		})
		return e.CM.Write(msg.ID, rFrame)
	} else {
		return nil
	}
}

func (e *Engine) handleSend(msg CnxMgrMsg, frame Frame) error {
	// TODO: destination validation
	// TODO: message ID creation
	newHeaders := make(map[string]string)
	for k, v := range frame.Headers {
		newHeaders[k] = v
	}

	dest, prs := frame.Headers["destination"]
	if !prs {
		return fmt.Errorf("error: client %s: no destination header", msg.ID)
	}

	messageFrame := Frame{
		Command: MESSAGE,
		Headers: newHeaders,
		Body:    frame.Body,
	}

	e.Store.Enqueue(dest, messageFrame)

	return nil
}
