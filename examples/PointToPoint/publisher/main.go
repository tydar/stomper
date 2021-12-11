package main

// publisher
import (
	"net"
	"time"
)

func main() {
	c, err := net.Dial("tcp", "queue:32801")
	for err != nil {
		c, err = net.Dial("tcp", "queue:32801")
	}
	defer c.Close()

	_, err = c.Write([]byte("CONNECT\naccept-version:1.2\nheart-beat:0,0\nhost:queue\n\n\000"))
	if err != nil {
		panic(err)
	}
	for {
		_, err = c.Write([]byte("SEND\ndestination:/queue/example\n\nTEST_MESSAGE\000")) 
		if err != nil {
			panic(err)
		}

		time.Sleep(10 * time.Second)
	}
}
