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
	handler := handleRequest

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

func handleRequest(w *response.Writer, req *request.Request) {
	if path, ok := strings.CutPrefix(req.RequestLine.RequestTarget, "/httpbin/"); ok {
		if err := proxyHandler(w, path); err != nil {
			log.Printf("proxy handler error: %v", err)
			_ = writeErrorResponse(w, response.InternalServerError, err)
		}
		return
	}

	if req.RequestLine.RequestTarget == "/video" {
		if err := videoHandler(w); err != nil {
			log.Printf("video handler error: %v", err)
			_ = writeErrorResponse(w, response.InternalServerError, err)
		}
		return
	}

	statusCode := response.Ok
	message := "Your request was an absolute banger."

	switch req.RequestLine.RequestTarget {
	case "/yourproblem":
		statusCode = response.BadRequest
		message = "Your request honestly kinda sucked."
	case "/myproblem":
		statusCode = response.InternalServerError
		message = "Okay, you know what? This one is on me."
	}

	if err := writeHTMLResponse(w, statusCode, message); err != nil {
		log.Printf("write response error: %v", err)
	}
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

func writeHTMLResponse(w *response.Writer, status response.StatusCode, msg string) error {
	body := generateResponseBody(status, msg)
	h := response.GetDefaultHeaders(len(body))
	h.Override("Content-Type", "text/html")

	if err := w.WriteStatusLine(status); err != nil {
		return err
	}

	if err := w.WriteHeaders(h); err != nil {
		return err
	}

	_, err := w.WriteBody(body)
	return err
}

func writeErrorResponse(w *response.Writer, status response.StatusCode, err error) error {
	if err == nil {
		return writeHTMLResponse(w, status, status.String())
	}

	return writeHTMLResponse(w, status, err.Error())
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

	if err := w.WriteStatusLine(response.Ok); err != nil {
		return err
	}

	if err := w.WriteHeaders(headrs); err != nil {
		return err
	}

	buf := make([]byte, 1024)
	fullResponse := []byte{}
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			if _, writeErr := w.WriteChunkedBody(buf[:n]); writeErr != nil {
				return writeErr
			}
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

	if _, err := w.WriteChunkedBodyDone(); err != nil {
		return err
	}

	if err := w.WriteTrailers(trailers); err != nil {
		return err
	}

	return nil
}

func videoHandler(w *response.Writer) error {
	file, err := os.ReadFile("./assets/vim.mp4")

	if err != nil {
		return err
	}

	h := response.GetDefaultHeaders(len(file))
	h.Override("Content-Type", "video/mp4")

	if err := w.WriteStatusLine(response.Ok); err != nil {
		return err
	}

	if err := w.WriteHeaders(h); err != nil {
		return err
	}

	_, err = w.WriteBody(file)
	return err
}
