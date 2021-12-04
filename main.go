package main

import "log"

func main() {
	comms := make(chan CnxMgrMsg)
	cm := NewConnectionManager("", 2000, comms)
	st := &MemoryStore{
		Queues: make(map[string][]Frame),
	}
	e := NewEngine(st, cm, comms)
	err := e.Start()
	if err != nil {
		log.Fatal(err)
	}
}
