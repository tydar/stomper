package main

import (
    "fmt"
    "testing"
    "reflect"
)

func TestScanNullTerm(t *testing.T) {
    empty := []byte("")
    nullTerm := []byte("Null-term\000")
    multipleNull := []byte("Null-term\000another\000")
    noNull := []byte("Justwords")
    var tests = []struct {
        input  []byte
        eof    bool
        adv    int
        token  []byte
        err    error
    }{
        {empty, false, 0, nil, nil},
        {nullTerm, false, 11, nullTerm, nil},
        {multipleNull, false, 11, nullTerm, nil},
        {noNull, false, 0, nil, nil},
        {noNull, true, 9, noNull, nil},
    }

    for _, tt := range tests {
        testname := fmt.Sprintf("%s", string(tt.input))
        t.Run(testname, func(t *testing.T) {
            advance, token, err := ScanNullTerm(tt.input, tt.eof)
            if err != tt.err {
                t.Errorf("err val: got %v wanted %v\n", err, tt.err)
            } else if advance != tt.adv {
                t.Errorf("wrong adv val: got %d wanted %d\n", advance, tt.adv)
            } else if !reflect.DeepEqual(token, tt.token) {
                t.Errorf("wrong token val: got %s wanted %s\n", string(token), string(tt.token))
            }
        })
    }
}
