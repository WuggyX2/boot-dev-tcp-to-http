package server

import (
	"fmt"
	"httpfromtcp/internal/request"
	"httpfromtcp/internal/response"
	"io"
)

type HandlerError struct {
	StatusCode response.StatusCode
	Message    string
}

func (herr HandlerError) Write(w io.Writer) error {
	_, err := fmt.Fprintf(w, "HTTP/1.1 %d %s\r\n", herr.StatusCode.Code(), herr.StatusCode.String())
	_, err = fmt.Fprint(w, "\r\n")
	_, err = fmt.Fprint(w, herr.Message)

	return err
}

type Handler func(w *response.Writer, req *request.Request)
