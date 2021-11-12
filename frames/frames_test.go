package frames

import (
    "fmt"
    "testing"
    "reflect"
)

func TestParseFrames(t *testing.T) {
    emptMap := make(map[string]string)
    var tests = []struct {
        raw  string
        want Frame
        err  bool // set to true if an error is expected
    }{
        {"SEND\n\n\000", Frame{Command: "SEND", Headers: emptMap, Body: ""}, false},
        {"SEND\ncontent-length:0\n\n\000", Frame{Command: "SEND", Headers: map[string]string{"content-length": "0"}, Body: "" }, false},
        {"SEND\n\n", Frame{}, true},
        {"BOOGIE\n\n\000", Frame{}, true},
        {"SEND\nbad header\n\000", Frame{}, true},
        {"SEND\ncontent-length:15\n\nabcd\000", Frame{}, true},
        {"SEND\ncontent-length:5\n\naaaaa\000", Frame{Command: "SEND", Headers: map[string]string{"content-length": "5"}, Body: "aaaaa"}, false},
    }

    for _, tt := range tests {
        testname := fmt.Sprintf("%s", tt.raw)
        t.Run(testname, func(t *testing.T) {
            parsed, err := ParseFrame(tt.raw)
            if (err != nil) && !tt.err {
                t.Errorf("got error: %s\n", err.Error())
            }
            if !reflect.DeepEqual(parsed, tt.want) {
                t.Errorf("got %+v / wanted %+v\n", parsed.Body, tt.want.Body)
            }
        })
    }

}
