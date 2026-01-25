package server

import (
	"fmt"
	response "httpfromtcp/internal/internal"
	"log"
	"net"
	"sync/atomic"
)

type Server struct {
	ln   net.Listener
	open atomic.Bool
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

	headers := response.GetDefaultHeaders(0)
	response.WriteStatusLine(conn, response.Ok)
	response.WriteHeaders(conn, headers)
}

func Serve(port int) (*Server, error) {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))

	if err != nil {
		return nil, err
	}

	server := Server{ln: ln, open: atomic.Bool{}}
	server.open.Store(true)
	go server.listen()

	return &server, nil
}
