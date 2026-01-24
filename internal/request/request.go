package request

import (
	"bytes"
	"errors"
	"httpfromtcp/internal/headers"
	"io"
	"slices"
	"strconv"
	"strings"
)

const crlf = "\r\n"
const supportedHTTPVersion = "1.1"
const bufferSize = 8

var validHTTPMethods = []string{"GET", "HEAD", "POST", "PUT", "DELETE", "CONNECT", "OPTIONS", "TRACE", "PATCH"}

type RequestState int

const (
	Initialized RequestState = iota
	ParsingHeaders
	ParsingBody
	Done
)

type Request struct {
	RequestLine RequestLine
	State       RequestState
	Headers     headers.Headers
	Body        []byte
	bodyReadInt int
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func (r *Request) parse(data []byte) (int, error) {
	totalBytesParsed := 0
	for r.State != Done {
		bytesRead, err := r.parseSingle(data[totalBytesParsed:])

		if err != nil {
			return 0, err
		}

		if bytesRead == 0 {
			break
		}

		totalBytesParsed += bytesRead
	}

	return totalBytesParsed, nil

}

func (r *Request) parseSingle(data []byte) (int, error) {
	switch r.State {
	case Initialized:
		bytesRead, reqLine, err := parseLineRequest(data)

		if err != nil {
			return 0, err
		}

		if bytesRead == 0 {
			return 0, nil
		}

		r.RequestLine = *reqLine
		r.State = ParsingHeaders
		return bytesRead, nil
	case ParsingHeaders:
		bytesRead, done, err := r.Headers.Parse(data)

		if err != nil {
			return 0, err
		}

		if done {
			r.State = ParsingBody
		}

		return bytesRead, nil

	case Done:
		return 0, errors.New("Trying to read data in done state")
	case ParsingBody:
		contentLength, ok := r.Headers.Get("Content-Length")

		if !ok {
			r.State = Done
			return len(data), nil
		}

		contentLen, err := strconv.Atoi(contentLength)

		if err != nil {
			return 0, errors.New("Cannot convert Content-Length header to a integer")
		}

		r.Body = slices.Concat(r.Body, data)
		r.bodyReadInt += len(data)

		if r.bodyReadInt > contentLen {
			return len(r.Body), errors.New("Request body is longer than stated in Content-Length header")
		}

		if r.bodyReadInt == contentLen {
			r.State = Done
		}

		return len(data), nil

	default:
		return 0, errors.New("Error: Unknown State")
	}
}

func RequestFromReader(reader io.Reader) (*Request, error) {

	buf := make([]byte, bufferSize)
	readToIndex := 0

	req := Request{
		State:   Initialized,
		Headers: headers.NewHeaders(),
		Body:    make([]byte, 0),
	}

	for req.State != Done {

		if len(buf) == readToIndex {
			newSize := len(buf) * 2
			newBuffer := make([]byte, newSize)
			copy(newBuffer, buf)
			buf = newBuffer
		}

		bytesRead, err := reader.Read(buf[readToIndex:])
		readToIndex += bytesRead

		if errors.Is(err, io.EOF) {
			if req.State != Done {
				return nil, errors.New("Incomplete request")
			}
			break
		}

		processedBytes, err := req.parse(buf[:readToIndex])

		if err != nil {
			return nil, err
		}

		if processedBytes > 0 {
			remaining := readToIndex - processedBytes
			copy(buf, buf[processedBytes:readToIndex])

			readToIndex = remaining
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
