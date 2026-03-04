package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"httpfromtcp/internal/request"
	"httpfromtcp/internal/response"
	"httpfromtcp/internal/server"
)

const port = 42069

func main() {

	handler := func(w *response.Writer, req *request.Request) {

		var message string
		var statusCode response.StatusCode

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
