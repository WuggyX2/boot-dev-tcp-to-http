package main

import (
	"bufio"
	"log"
	"net"
	"os"
)

func main() {
	udpAddress, err := net.ResolveUDPAddr("udp", "localhost:42069")

	if err != nil {
		log.Fatal("could not resolve udp address localhost:42069")
	}

	udpConnection, err := net.DialUDP("udp", nil, udpAddress)

	if err != nil {
		log.Fatal("could not establish a udp connectuon to localhost:42069")
		log.Fatal("could not establish a udp connectuon to localhost:42069")
	}

	reader := bufio.NewReader(os.Stdin)

	for {
		log.Println(">")
		input, err := reader.ReadString('\n')

		if err != nil {
			log.Println("Error occured: " + err.Error())
			break
		}

		_, err = udpConnection.Write([]byte(input))

		if err != nil {
			log.Fatal("Error writing to udp connection")
		}

	}
}
