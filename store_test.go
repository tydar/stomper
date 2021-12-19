package main

import (
	"fmt"
	"reflect"
	"testing"
)

func TestMemoryStoreEnqueue(t *testing.T) {
	emptMap := make(map[string][]Frame)
	testDest := map[string][]Frame{
		"/queue/test": make([]Frame, 0),
	}
	emptHead := make(map[string]string)
	fr := Frame{Command: "SEND", Headers: emptHead, Body: ""}
	var tests = []struct {
		initial map[string][]Frame
		dest    string
		frame   Frame
		final   map[string][]Frame
		err     bool
	}{
		{
			initial: emptMap,
			dest:    "/queue/test",
			frame:   fr,
			final:   map[string][]Frame{"/queue/test": {fr}},
			err:     true,
		},
		{
			initial: testDest,
			dest:    "/queue/test",
			frame:   fr,
			final:   map[string][]Frame{"/queue/test": {fr}},
			err:     false,
		},
	}
	for _, tt := range tests {
		testname := fmt.Sprintf("%s", tt.initial)
		t.Run(testname, func(t *testing.T) {
			ms := MemoryStore{Queues: tt.initial}
			err := ms.Enqueue(tt.dest, tt.frame)
			if (err != nil) && tt.err {
				t.Log("Received expected error!")
			} else if (err == nil) && tt.err {
				t.Errorf("Expected an error!\n")
			} else if !reflect.DeepEqual(ms.Queues, tt.final) {
				t.Errorf("got %+v / wanted %+v\n", ms.Queues, tt.final)
			}
		})
	}
}

func TestMemoryStorePop(t *testing.T) {
	emptHead := make(map[string]string)
	fr := Frame{Command: "SEND", Headers: emptHead, Body: ""}
	emptMap := make(map[string][]Frame)
	mapWithOne := map[string][]Frame{
		"/queue/test": {fr},
	}
	destWithNone := map[string][]Frame{
		"/queue/test": make([]Frame, 0),
	}
	var tests = []struct {
		initial map[string][]Frame
		dest    string
		result  Frame
		err     bool
	}{
		{
			initial: emptMap,
			dest:    "/queue/test",
			result:  Frame{},
			err:     true,
		},
		{
			initial: mapWithOne,
			dest:    "/queue/test",
			result:  fr,
			err:     false,
		},
		{
			initial: destWithNone,
			dest:    "/queue/test",
			result:  Frame{},
			err:     true,
		},
	}
	for _, tt := range tests {
		testname := fmt.Sprintf("%s", tt.initial)
		t.Run(testname, func(t *testing.T) {
			ms := MemoryStore{Queues: tt.initial}
			frRes, err := ms.Pop(tt.dest)
			t.Log("Entering testing conditionals")
			if (err != nil) && tt.err {
				t.Log("Received expected error!")
			} else if (err == nil) && tt.err {
				t.Errorf("Expected an error!\n")
			} else if !reflect.DeepEqual(frRes, tt.result) {
				t.Errorf("got %+v / wanted %+v\n", frRes, tt.result)
			}
		})
	}
}

func TestMemoryStoreAddPrs(t *testing.T) {
	emptMap := make(map[string][]Frame)
	ms := MemoryStore{Queues: emptMap}

	err := ms.AddDestination("/queue/main")
	if err != nil {
		t.Error("destination add error: ", err)
	}

	dest := ms.Destinations()
	if dest[0] != "/queue/main" {
		t.Errorf("destination not added correctly: got %s wanted %s", dest, "/queue/main")
	}

	err = ms.AddDestination("/queue/main")
	if err == nil {
		t.Error("allowed duplicate destination creation")
	}

}
