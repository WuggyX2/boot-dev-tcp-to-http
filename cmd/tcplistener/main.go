package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
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
		channel := getLinesChannel(conn)
		for msg := range channel {
			fmt.Println(msg)
		}
		fmt.Println("Connection has been Closed")
	}

}

func getLinesChannel(f io.ReadCloser) <-chan string {
	linesChannel := make(chan string)

	go func() {
		defer f.Close()
		defer close(linesChannel)
		buffer := make([]byte, 8)
		var currentLine string
		for {
			n, err := f.Read(buffer)

			if err != nil {
				if currentLine != "" {
					linesChannel <- currentLine
				}

				if errors.Is(err, io.EOF) {
					break
				}

				fmt.Printf("error: %s\n", err.Error())
				break
			}

			bufferAsString := string(buffer[:n])
			parts := strings.Split(bufferAsString, "\r\n")

			currentLine += parts[0]

			if len(parts) == 2 {
				linesChannel <- currentLine
				currentLine = parts[1]
			}
		}
	}()

	return linesChannel
}
