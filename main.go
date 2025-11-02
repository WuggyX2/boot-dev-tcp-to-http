package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

func main() {
	file, err := os.Open("messages.txt")

	if err != nil {
		log.Fatal("Error opening messages file")
	}

	channel := getLinesChannel(file)

	for msg := range channel {
		fmt.Printf("read: %s\n", msg)
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
			_, err := f.Read(buffer)

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

			bufferAsString := fmt.Sprintf("%s", buffer)
			parts := strings.Split(bufferAsString, "\n")

			currentLine += parts[0]

			if len(parts) == 2 {
				linesChannel <- currentLine
				currentLine = parts[1]
			}
		}
	}()

	return linesChannel
}
