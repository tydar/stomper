package main

// subscriber
import (
	"net"
	"io"
	"os"
	"fmt"
)

func main() {
	c, err := net.Dial("tcp", "queue:32801")
	for err != nil {
		c, err = net.Dial("tcp", "queue:32801")
	}
	defer c.Close()

	f, err := os.OpenFile("./messages.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}

	_, err = c.Write([]byte("CONNECT\naccept-version:1.2\nheart-beat:0,0\nhost:queue\n\n\000"))
	if err != nil {
		panic(err)
	}

	_, err = c.Write([]byte("SUBSCRIBE\nid:1\ndestination:/queue/example\n\n\000"))

	_, err = io.Copy(f, c)
	if err != nil {
		fmt.Println(err)
	}
}
