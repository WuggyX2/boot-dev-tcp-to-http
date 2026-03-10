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
	if _, err := fmt.Fprintf(w, "HTTP/1.1 %d %s\r\n", herr.StatusCode.Code(), herr.StatusCode.String()); err != nil {
		return err
	}

	if _, err := fmt.Fprint(w, "\r\n"); err != nil {
		return err
	}

	if _, err := fmt.Fprint(w, herr.Message); err != nil {
		return err
	}

	return nil
}

type Handler func(w *response.Writer, req *request.Request)
