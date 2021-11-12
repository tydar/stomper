package frames

import (
    "strings"
    "errors"
    "strconv"
    "bytes"
    "log"
)

const (
    CONNECT = "CONNECT"
    STOMP = "STOMP"
    CONNECTED = "CONNECTED"
    SEND = "SEND"
    SUBSCRIBE = "SUBSCRIBE"
    UNSUBSCRIBE = "UNSUBSCRIBE"
    ACK = "ACK"
    NACK = "NACK"
    BEGIN = "BEGIN"
    COMMIT = "COMMIT"
    ABORT = "ABORT"
    DISCONNECT = "DISCONNECT"
    MESSAGE = "MESSAGE"
    RECEIPT = "RECEIPT"
    ERROR = "ERROR"
)

type Frame struct {
    Command string
    Headers map[string]string
    Body    string
}

func ParseFrame(text string) (Frame, error) {
    tokens := strings.Split(text, "\n")
    if len(tokens) < 3 {
        return Frame{}, errors.New("Invalid frame, too few newlines per STOMP specification.")
    }

    // first line should be the command
    command := tokens[0]

    if !validateCommand(command) {
        return Frame{}, errors.New("Invalid command.")
    }

    // an arbitrary number of lines will be headers
    // terminated by a blank line due to double \n characters
    current := 1
    stillHeaders := true
    headers := make(map[string]string)

    for stillHeaders {
        // test if we have reached a blank line signifying the end of the headers
        if len(tokens[current]) > 0 {
            k, v, err := parseHeader(tokens[current])
            if err != nil {
                return Frame{}, err
            }
            _, prs := headers[k] // if there is a duplicate header, drop each occurance after the first
            if !prs {
                headers[k] = v
            }
            current++
        }  else {
            current++
            stillHeaders = false
        }
    }

    possibleBody := strings.Join(tokens[current:], "\n")
    log.Printf("%s\n", possibleBody)
    cl, prs := headers["content-length"]
    contentLength := -1
    var errConv error
    if prs {
        contentLength, errConv = strconv.Atoi(cl)
        if errConv != nil {
            return Frame{}, errConv
        }
    }

    if prs && contentLength >= 0 {
        // if we have a content-length header with value greater than 0,
        // we want to ensure we received the full header
        // so we check against len(possibleBody) - 1 to count out a potential null byte
        // required for termination
        if (len(possibleBody) - 1) < contentLength {
            return Frame{}, errors.New("Incomplete frame.")
        }
        // now that we know we have a long enough frame body
        // we can slice it at content-length
        // and check that it was null terminated / not too long
        var rest []byte
        rest = []byte(possibleBody[contentLength:])
        if len(rest) == 0 || rest[0] != '\000' {
            return Frame{}, errors.New("Termination error.")
        }
        possibleBody = possibleBody[:contentLength]
    } else {
        // if we have no content-length header
        // we need to ensure there is a null byte
        byteBody := []byte(possibleBody)
        terminatorIndex := bytes.IndexRune(byteBody, '\000')
        if terminatorIndex < 0 {
            return Frame{}, errors.New("Termination error.")
        } else {
            // and if there was, slice the body at the first null byte
            possibleBody = possibleBody[:terminatorIndex]
        }
    }
    
    return Frame{
        Command: command,
        Headers: headers,
        Body:    possibleBody,
    }, nil
}

func validateCommand(cmd string) bool {
    switch cmd {
        case CONNECT, STOMP, CONNECTED, SEND, SUBSCRIBE, UNSUBSCRIBE, ACK,
             NACK, BEGIN, COMMIT, ABORT, DISCONNECT, MESSAGE, RECEIPT, ERROR:
            return true
        default:
            return false
    }
}

func parseHeader(header string) (string, string, error) {
    // parse one header line into a key and a vlue
    tokens := strings.SplitN(header, ":", 2) // only want to split on the first :
    if len(tokens) != 2 {
        return "", "", errors.New("Malformed header.")
    }
    return tokens[0], tokens[1], nil
}
