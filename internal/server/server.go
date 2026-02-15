package server

import (
	"bytes"
	"fmt"
	"httpfromtcp/internal/request"
	"httpfromtcp/internal/response"
	"log"
	"net"
	"sync/atomic"
)

type Server struct {
	ln      net.Listener
	open    atomic.Bool
	handler Handler
}

func (s *Server) Close() error {
	if s.ln == nil {
		return nil
	}

	s.open.Store(false)
	return s.ln.Close()
}

func (s *Server) listen() {

	for {
		conn, err := s.ln.Accept()
		if err != nil {
			if !s.open.Load() {
				return
			}
			log.Printf("Error accepting connection: %v", err)
			continue
		}
		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()

	req, err := request.RequestFromReader(conn)

	if err != nil {
		hErr := &HandlerError{
			StatusCode: response.BadRequest,
			Message:    err.Error(),
		}
		hErr.Write(conn)
		return
	}

	buf := bytes.NewBuffer([]byte{})
	hErr := s.handler(buf, req)

	if hErr != nil {
		hErr.Write(conn)
		return
	}

	body := buf.Bytes()
	headers := response.GetDefaultHeaders(len(body))
	response.WriteStatusLine(conn, response.Ok)
	response.WriteHeaders(conn, headers)
	conn.Write(body)
}

func Serve(port int, handler Handler) (*Server, error) {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))

	if err != nil {
		return nil, err
	}

	server := Server{ln: ln, open: atomic.Bool{}, handler: handler}
	server.open.Store(true)
	go server.listen()

	return &server, nil
}
