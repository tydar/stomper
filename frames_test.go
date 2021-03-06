package main

import (
	"fmt"
	"reflect"
	"testing"
)

func TestParseFrames(t *testing.T) {
	emptMap := make(map[string]string)
	var tests = []struct {
		raw  string
		want Frame
		err  bool // set to true if an error is expected
	}{
		{"SEND\n\n\000", Frame{Command: "SEND", Headers: emptMap, Body: ""}, false},
		{"SEND\ncontent-length:0\n\n\000", Frame{Command: "SEND", Headers: map[string]string{"content-length": "0"}, Body: ""}, false},
		{"SEND\n\n", Frame{}, true},
		{"BOOGIE\n\n\000", Frame{}, true},
		{"SEND\nbad header\n\000", Frame{}, true},
		{"SEND\ncontent-length:15\n\nabcd\000", Frame{}, true},
		{"SEND\ncontent-length:5\n\naaaaa\000", Frame{Command: "SEND", Headers: map[string]string{"content-length": "5"}, Body: "aaaaa"}, false},
	}

	for _, tt := range tests {
		testname := tt.raw
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

func TestUnmarshalFrame(t *testing.T) {
	emptMap := make(map[string]string)
	var tests = []struct {
		f    Frame
		want string
	}{
		{Frame{Command: "SEND", Headers: emptMap, Body: ""}, "SEND\n\n\000"},
		{Frame{Command: "SEND", Headers: map[string]string{"content-length": "5"}, Body: "aaaaa"}, "SEND\ncontent-length:5\n\naaaaa\000"},
	}

	for _, tt := range tests {
		testname := fmt.Sprintf("%+v", tt.f)
		t.Run(testname, func(t *testing.T) {
			unmarshalled := UnmarshalFrame(tt.f)
			if unmarshalled != tt.want {
				t.Errorf("got %s / wanted %s\n", unmarshalled, tt.want)
			}
		})
	}
}
