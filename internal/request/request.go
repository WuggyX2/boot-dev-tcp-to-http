package request

import (
	"bytes"
	"errors"
	"io"
	"slices"
	"strings"
)

const crlf = "\r\n"
const supportedHTTPVersion = "1.1"
const bufferSize = 8

var validHTTPMethods = []string{"GET", "HEAD", "POST", "PUT", "DELETE", "CONNECT", "OPTIONS", "TRACE", "PATCH"}

type RequestState int

const (
	Initialized RequestState = iota
	Done
)

type Request struct {
	RequestLine RequestLine
	State       RequestState
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func (r *Request) parse(data []byte) (int, error) {
	if r.State == Done {
		return 0, errors.New("Trying to read data in done state")
	}

	if r.State == Initialized {
		bytesRead, reqLine, err := parseLineRequest(data)

		if err != nil {
			return 0, err
		}

		if bytesRead == 0 {
			return 0, nil
		}

		r.RequestLine = *reqLine
		r.State = Done

		return bytesRead, nil
	}

	return 0, errors.New("Error: Unknown State")
}

func RequestFromReader(reader io.Reader) (*Request, error) {

	buf := make([]byte, bufferSize)
	readToIndex := 0

	req := Request{
		State: Initialized,
	}

	for req.State != Done {

		if len(buf) == readToIndex {
			newSize := len(buf) * 2
			newBuffer := make([]byte, newSize)
			copy(newBuffer, buf)
			buf = newBuffer
		}

		bytesRead, readErr := reader.Read(buf[readToIndex:])
		readToIndex += bytesRead

		processedBytes, err := req.parse(buf[:readToIndex])

		if err != nil {
			return nil, err
		}

		if processedBytes > 0 {
			remaining := readToIndex - processedBytes
			copy(buf, buf[processedBytes:readToIndex])

			readToIndex = remaining
		}

		if readErr == io.EOF {
			req.State = Done
			break
		}
	}

	return &req, nil
}

func parseLineRequest(input []byte) (int, *RequestLine, error) {
	idx := bytes.Index(input, []byte(crlf))
	if idx == -1 {
		return 0, nil, nil
	}

	bytesRead := idx + len(crlf)

	requestLineText := string(input[:idx])
	reqLine, err := validateAndCreateRequestLine(requestLineText)

	if err != nil {
		return bytesRead, nil, err
	}

	return bytesRead, reqLine, nil
}

func validateAndCreateRequestLine(input string) (*RequestLine, error) {
	parts := strings.Split(input, " ")

	if len(parts) != 3 {
		return nil, errors.New("Invalid number of parts in request line.")
	}

	if !slices.Contains(validHTTPMethods, parts[0]) {
		return nil, errors.New("The given http method is not valid")
	}

	httpVersionParts := strings.Split(parts[2], "/")

	if len(httpVersionParts) != 2 {
		return nil, errors.New("HTTP version in request line is not correctly formatted. Should be HTTP/{version number}.")
	}

	if httpVersionParts[1] != supportedHTTPVersion {
		return nil, errors.New("HTTP version is not supported")
	}

	reqLine := RequestLine{
		Method:        parts[0],
		RequestTarget: parts[1],
		HttpVersion:   httpVersionParts[1],
	}
	return &reqLine, nil
}
