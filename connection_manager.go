package main

import (
    "bytes"
)

// Custom scanner to split incoming stream on \000
func ScanNullTerm(data []byte, atEOF bool) (int, []byte, error) {
    // if we're at EOF, we're done for now
    if atEOF && len(data) == 0 {
        return 0, nil, nil
    }

    // if we find a '\000', return the data up to and including that index
    if i := bytes.IndexByte(data, '\000'); i >= 0 {
        // there is a null-terminated frame
        // we want to return it, including the null byte
        return i+2, data[0:i+1], nil
    }

    // if we are at EOF and we have data, return it so we can see what's going on
    if atEOF {
        return len(data), data, nil
    }
    return 0, nil, nil
}
