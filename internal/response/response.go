package response

import (
	"errors"
	"fmt"
	"httpfromtcp/internal/headers"
	"io"
	"strconv"
)

type StatusCode int
type WriterState int

const (
	Ok StatusCode = iota
	BadRequest
	InternalServerError
)

const (
	WritingStatusLine = iota
	WritingHeaders
	WritingBody
)

type Writer struct {
	writer      io.Writer
	writerState WriterState
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
	if w.writerState != WritingStatusLine {
		return errors.New("Writer state is not WritingStatusLine, call methods in the correct order")
	}
	_, err := fmt.Fprintf(w.writer, "HTTP/1.1 %d %s\r\n", statusCode.Code(), statusCode.String())

	if err == nil {
		w.writerState = WritingHeaders
	}

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
	if w.writerState != WritingHeaders {
		return errors.New("Writer state is not WritingHeaders, call methods in the correct order")
	}

	for key, val := range headers {
		_, err := fmt.Fprintf(w.writer, "%s: %s\r\n", key, val)

		if err != nil {
			return err
		}
	}

	_, err := fmt.Fprint(w.writer, "\r\n")

	if err == nil {
		w.writerState = WritingBody
	}

	return err
}

func (w *Writer) WriteBody(p []byte) (int, error) {
	if w.writerState != WritingBody {
		return 0, errors.New("Writer state is not WritingBody, call methods in the correct order")
	}

	bytesWritten, err := w.writer.Write(p)

	if err != nil {
		return 0, err
	}

	return bytesWritten, nil
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{writer: w, writerState: WritingStatusLine}
}
