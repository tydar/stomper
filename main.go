package main

import (
    "fmt"
    "reflect"
    "github.com/tydar/stomper/frames"
)

func main() {
    f, err := frames.ParseFrame("SEND\n\n\000")
    f2 := frames.Frame{Command: "SEND", Headers: make(map[string]string), Body: ""}
    if err != nil {
        panic(err)
    }
    fmt.Println(reflect.DeepEqual(f, f2))
}
