package main

import (
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"httpfromtcp/internal/headers"
	"httpfromtcp/internal/request"
	"httpfromtcp/internal/response"
	"httpfromtcp/internal/server"
)

const port = 42069

func main() {

	handler := func(w *response.Writer, req *request.Request) {

		var message string
		var statusCode response.StatusCode

		if path, ok := strings.CutPrefix(req.RequestLine.RequestTarget, "/httpbin/"); ok {
			err := proxyHandler(w, path)

			if err != nil {
				headers := response.GetDefaultHeaders(0)
				w.WriteStatusLine(response.InternalServerError)
				w.WriteHeaders(headers)
				w.WriteBody([]byte(err.Error()))
			}
			return
		}

		switch req.RequestLine.RequestTarget {
		case "/yourproblem":
			statusCode = response.BadRequest
			message = "Your request honestly kinda sucked."
		case "/myproblem":
			statusCode = response.InternalServerError
			message = "Okay, you know what? This one is on me."
		default:
			statusCode = response.Ok
			message = "Your request was an absolute banger."
		}

		body := generateResponseBody(statusCode, message)
		headers := response.GetDefaultHeaders(len(body))
		headers.Override("Content-Type", "text/html")

		w.WriteStatusLine(statusCode)
		w.WriteHeaders(headers)
		w.WriteBody(body)

	}

	server, err := server.Serve(port, handler)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}

func generateResponseBody(status response.StatusCode, msg string) []byte {
	body := fmt.Sprintf(`<html>
  <head>
    <title>%d %s</title>
  </head>
  <body>
    <h1>%s</h1>
    <p>%s</p>
  </body>
</html>`, status.Code(), status.String(), status.String(), msg)

	return []byte(body)

}

func proxyHandler(w *response.Writer, path string) error {
	headrs := response.GetDefaultHeadersForChunked()
	headrs.Set("Trailer", "X-Content-SHA256")
	headrs.Set("Trailer", "X-Content-Length")

	resp, err := http.Get("https://httpbin.org/" + path)

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	w.WriteStatusLine(response.Ok)
	w.WriteHeaders(headrs)

	buf := make([]byte, 1024)
	fullResponse := []byte{}
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			w.WriteChunkedBody(buf[:n])
			fullResponse = append(fullResponse, buf[:n]...)
		}
		if err == io.EOF {
			break
		}

		if err != nil {
			return err
		}
	}

	hash := sha256.Sum256(fullResponse)

	trailers := headers.NewHeaders()

	trailers.Set("X-Content-SHA256", fmt.Sprintf("%x", hash))
	trailers.Set("X-Content-Length", strconv.Itoa(len(fullResponse)))

	w.WriteChunkedBodyDone()
	w.WriteTrailers(trailers)

	return nil
}
