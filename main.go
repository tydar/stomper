package main

import (
	"fmt"
	"reflect"
)

func main() {
	f, err := ParseFrame("SEND\n\n\000")
	f2 := Frame{Command: "SEND", Headers: make(map[string]string), Body: ""}
	if err != nil {
		panic(err)
	}
	fmt.Println(reflect.DeepEqual(f, f2))
}
