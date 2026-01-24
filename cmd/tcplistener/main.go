package main

import (
	"fmt"
	"httpfromtcp/internal/request"
	"log"
	"net"
)

func main() {
	tcpListener, err := net.Listen("tcp", "127.0.0.1:42069")

	if err != nil {
		log.Fatal("Error when trying to listen tcp on localhost post 42069")
	}

	defer tcpListener.Close()
	for {
		conn, err := tcpListener.Accept()

		if err != nil {
			log.Fatal("Error trying to Accept a tcp connection")
		}

		fmt.Println("Connection has been Accepted")
		req, err := request.RequestFromReader(conn)

		if err != nil {
			log.Fatal("Error occured when parsing requestline")
		}

		fmt.Printf(
			"Request line:\n- Method: %s\n- Target: %s\n- Version: %s\n",
			req.RequestLine.Method,
			req.RequestLine.RequestTarget,
			req.RequestLine.HttpVersion,
		)

		fmt.Println("Headers:")
		for key, val := range req.Headers {
			fmt.Printf("- %s: %s\n", key, val)
		}

		fmt.Println("Body:")
		fmt.Println(string(req.Body))

		fmt.Println("Connection has been Closed")
	}

}
