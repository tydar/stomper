package main

import (
	"fmt"
	"log"
	"strconv"
)

type Engine struct {
	CM            *ConnectionManager
	SM            *SubscriptionManager
	TM            *TransactionManager
	MS            *MetricsService
	metricsServer bool
	msAddr        string
	Incoming      chan CnxMgrMsg
	Store         Store
	SendWorkers   int
}

func NewEngine(st Store, cm *ConnectionManager, inc chan CnxMgrMsg, sendWorkers int, metricsServer bool, msAddr string) *Engine {
	return &Engine{
		CM:            cm,
		Store:         st,
		Incoming:      inc,
		SM:            NewSubscriptionManager(),
		TM:            NewTransactionManager(),
		SendWorkers:   sendWorkers,
		MS:            NewMetricsService(),
		metricsServer: metricsServer,
		msAddr:        msAddr,
	}
}

func (e *Engine) Start() error {
	err := e.CM.Start()
	if err != nil {
		return err
	}

	// start send workers
	go e.WorkerManager(e.SendWorkers)

	// if the metrics server flag is true
	if e.metricsServer {
		go e.MS.ListenAndServeJSON(e.msAddr)
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
			} else {
				e.MS.IncReceived()
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
				err = e.handleDisconnect(msg, frame)
				if err != nil {
					log.Println(err)
				} else {
					err = e.handleReceipt(msg, frame)
					if err != nil {
						log.Printf("ERROR: client %s write error: %s\n", msg.ID, err)
					}
				}
			case BEGIN:
				err = e.handleBegin(msg, frame)
				if err != nil {
					log.Println(err)
					err := e.handleError(msg, err)
					if err != nil {
						log.Printf("ERROR: client %s write error: %s\n", msg.ID, err)
					}
				}
			case ABORT:
				err = e.handleAbort(msg, frame)
				if err != nil {
					log.Println(err)
					err := e.handleError(msg, err)
					if err != nil {
						log.Printf("ERROR: client %s write error: %s\n", msg.ID, err)
					}
				}
			case COMMIT:
				err = e.handleCommit(msg, frame)
				if err != nil {
					err := fmt.Errorf("handleCommit: %v", err)
					log.Println(err)
					err = e.handleError(msg, err)
					if err != nil {
						log.Printf("handleError: client %s: %v\n", msg.ID, err)
					}
				}
			}
		} else if msg.Type == CONNECTION_CLOSED {
			e.SM.UnsubscribeAll(msg.ID)
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

func (e *Engine) handleDisconnect(msg CnxMgrMsg, frame Frame) error {
	return e.CM.Disconnect(msg.ID)
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
	create, prs := frame.Headers["create"]
	if !prs {
		create = "false"
	}

	destPrs := e.Store.Prs(dest)
	if create != "true" {
		if !destPrs {
			return fmt.Errorf("error: no such destination %s", dest)
		}
	} else {
		if !destPrs {
			err := e.Store.AddDestination(dest)
			if err != nil {
				return err
			}
		}
	}
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
	e.MS.IncError()
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

	// if we have a
	tx, prs := frame.Headers["transaction"]
	if prs {
		txFrame := Frame{
			Command: SEND,
			Headers: newHeaders,
			Body:    frame.Body,
		}
		return e.TM.AddFrame(tx, msg.ID, txFrame)
	}

	messageFrame := prepareMessage(frame)

	return e.Store.Enqueue(dest, messageFrame)
}

func (e *Engine) handleBegin(msg CnxMgrMsg, frame Frame) error {
	txId, ok := frame.Headers["transaction"]
	if !ok {
		return fmt.Errorf("error: client %s: no transaction header", msg.ID)
	}

	return e.TM.StartTransaction(txId, msg.ID)
}

func (e *Engine) handleAbort(msg CnxMgrMsg, frame Frame) error {
	txId, ok := frame.Headers["transaction"]
	if !ok {
		return fmt.Errorf("error: client %s: no transaction header", msg.ID)
	}

	return e.TM.AbortTransaction(txId, msg.ID)
}

func (e *Engine) handleCommit(msg CnxMgrMsg, frame Frame) error {
	// 1) receive transaction
	// 2) Process all frames in transaction to MESSAGE
	// 3) Enqueue Tx []Frame with e.Store.EnqueueTx
	txId, ok := frame.Headers["transaction"]
	if !ok {
		return fmt.Errorf("commit %s: no transaction header", msg.ID)
	}

	tx, err := e.TM.CommitTransaction(txId, msg.ID)
	if err != nil {
		return fmt.Errorf("CommitTransaction: %v", err)
	}

	finalTx := make(map[string]Frame, len(tx.frames))
	for i := range tx.frames {
		newFr := prepareMessage(tx.frames[i])
		dest := newFr.Headers["destination"] // should be guaranteed by initial handleSend call
		finalTx[dest] = newFr
	}

	return e.Store.EnqueueTx(finalTx)
}

// deep copy a SEND frame to a message frame to avoid race conditions
func prepareMessage(frame Frame) Frame {
	newHeaders := make(map[string]string)
	for k, v := range frame.Headers {
		newHeaders[k] = v
	}

	return Frame{
		Command: MESSAGE,
		Headers: newHeaders,
		Body:    frame.Body,
	}
}

// msg is of type []Frame
// if len(msg) == 1 that means we're sending a regular message
// if len(msg) > 1 that means we're working with a transaction
type SendJob struct {
	msg           []Frame
	subscriptions []Subscription
}

func (e *Engine) SendWorker(id int, jobs <-chan SendJob) {
	for j := range jobs {
		if len(j.msg) == 1 {
			for _, sub := range j.subscriptions {
				msg := j.msg[0]
				clientID := sub.ClientID
				uniqueHeaders := make(map[string]string)
				for k, v := range msg.Headers {
					uniqueHeaders[k] = v
				}
				uniqueHeaders["subscription"] = sub.ID

				uFrame := Frame{
					Command: msg.Command,
					Headers: uniqueHeaders,
					Body:    msg.Body,
				}
				uFrString := UnmarshalFrame(uFrame)
				err := e.CM.Write(clientID, uFrString)
				if err != nil {
					log.Printf("worker %d: SEND_ERROR: %s\n", id, err)
					e.MS.IncError()
				} else {
					e.MS.IncSent()
				}
			}
		}
	}
}

func (e *Engine) WorkerManager(numWorkers int) {
	log.Printf("starting %d workers\n", numWorkers)

	sChan := make(chan SendJob)
	for i := 0; i < numWorkers; i++ {
		go e.SendWorker(i, sChan)
	}
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
					sChan <- SendJob{msg: messageFrame, subscriptions: subscribers}
				}
			}
		}
	}
}
