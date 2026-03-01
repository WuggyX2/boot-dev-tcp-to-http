package response

import (
	"fmt"
	"httpfromtcp/internal/headers"
	"io"
	"strconv"
)

type StatusCode int

const (
	Ok StatusCode = iota
	BadRequest
	InternalServerError
)

type Writer struct {
	writer io.Writer
}

func (s StatusCode) Code() int {
	switch s {
	case Ok:
		return 200
	case BadRequest:
		return 400
	case InternalServerError:
		return 500
	default:
		return 500
	}
}

func (s StatusCode) String() string {

	switch s {
	case Ok:
		return "OK"
	case BadRequest:
		return "Bad Request"
	case InternalServerError:
		return "Internal Server Error"
	default:
		return ""
	}
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	_, err := fmt.Fprintf(w.writer, "HTTP/1.1 %d %s\r\n", statusCode.Code(), statusCode.String())
	return err
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	headers := headers.NewHeaders()

	headers.Set("Content-Length", strconv.Itoa(contentLen))
	headers.Set("Connection", "close")
	headers.Set("Content-Type", "text/plain")

	return headers
}

func (w *Writer) WriteHeaders(headers headers.Headers) error {
	for key, val := range headers {
		_, err := fmt.Fprintf(w.writer, "%s: %s\r\n", key, val)

		if err != nil {
			return err
		}
	}

	_, err := fmt.Fprint(w.writer, "\r\n")
	return err
}

func (w *Writer) WriteBody(p []byte) (int, error) {

	return 0, nil
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{writer: w}
}
