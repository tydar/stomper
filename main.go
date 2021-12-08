package main

import "log"

func main() {
	comms := make(chan CnxMgrMsg)
	cm := NewConnectionManager("", 2000, comms, 10)
	st := &MemoryStore{
		Queues: map[string][]Frame{"/queue/test": make([]Frame, 0)},
	}
	e := NewEngine(st, cm, comms)
	err := e.Start()
	if err != nil {
		log.Fatal(err)
	}
}
