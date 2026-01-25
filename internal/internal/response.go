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

func WriteStatusLine(w io.Writer, statusCode StatusCode) error {
	_, err := fmt.Fprintf(w, "HTTP/1.1 %d %s\r\n", statusCode.Code(), statusCode.String())
	return err
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	headers := headers.NewHeaders()

	headers.Set("Content-Length", strconv.Itoa(contentLen))
	headers.Set("Connection", "close")
	headers.Set("Content-Type", "text/plain")

	return headers
}

func WriteHeaders(w io.Writer, headers headers.Headers) error {
	for key, val := range headers {
		_, err := fmt.Fprintf(w, "%s: %s\r\n", key, val)

		if err != nil {
			return err
		}
	}

	_, err := fmt.Fprint(w, "\r\n")
	return err
}
