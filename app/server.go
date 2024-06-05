package main

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
)

func handle(conn net.Conn) {
	request, err := http.ReadRequest(bufio.NewReader(conn))
	if err != nil {
		fmt.Println("Error reading request.", err.Error())
	}
	fmt.Printf("Request: %s %s \n", request.Method, request.URL)
	toEcho, isEcho := strings.CutPrefix(request.URL.Path, "/echo/")
	if isEcho {
		fmt.Fprintf(conn, "HTTP/1.1 200 OK\r\n")
		fmt.Fprintf(conn, "Content-Type: text/plain\r\n")
		fmt.Fprintf(conn, "Content-Length: %d\r\n", len(toEcho))
		fmt.Fprintf(conn, "\r\n") //headers
		fmt.Fprintf(conn, "%s\r\n", toEcho)
		return
	}
	if request.URL.Path == "/" {
		fmt.Fprintf(conn, "HTTP/1.1 200 OK\r\n\r\n")
		return
	}
	fmt.Fprintf(conn, "HTTP/1.1 404 Not Found\r\n\r\n")
}

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")

	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	conn, err := l.Accept()
	if err != nil {
		fmt.Println("Error accepting connection: ", err.Error())
		os.Exit(1)

	}
	defer conn.Close()
	handle(conn)
}
