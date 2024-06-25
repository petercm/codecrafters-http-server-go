package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
)

func okContent(conn net.Conn, content string) {
	fmt.Fprintf(conn, "HTTP/1.1 200 OK\r\n")
	fmt.Fprintf(conn, "Content-Type: text/plain\r\n")
	fmt.Fprintf(conn, "Content-Length: %d\r\n", len(content))
	fmt.Fprintf(conn, "\r\n") //headers
	fmt.Fprintf(conn, "%s\r\n", content)
}

func okFile(conn net.Conn, content []byte) {
	fmt.Fprintf(conn, "HTTP/1.1 200 OK\r\n")
	fmt.Fprintf(conn, "Content-Type: application/octet-stream\r\n")
	fmt.Fprintf(conn, "Content-Length: %d\r\n", len(content))
	fmt.Fprintf(conn, "\r\n") //headers
	fmt.Fprintf(conn, "%s\r\n", content)
}

func handle(conn net.Conn) {
	defer conn.Close()
	request, err := http.ReadRequest(bufio.NewReader(conn))
	if err != nil {
		fmt.Println("Error reading request.", err.Error())
	}
	fmt.Printf("Request: %s %s \n", request.Method, request.URL)

	toEcho, isEcho := strings.CutPrefix(request.URL.Path, "/echo/")
	if isEcho {
		okContent(conn, toEcho)
		return
	}

	fileName, isFile := strings.CutPrefix(request.URL.Path, "/files/")
	if isFile {
		if request.Method == "POST" {
			contentLen := request.ContentLength
			if contentLen <= 0 {
				panic("content length less than 0")
			}
			bodyContent := make([]byte, contentLen)
			_, err = io.ReadFull(request.Body, bodyContent)
			if err != nil {
				panic(err)
			}
			err = os.WriteFile(os.Args[2]+fileName, bodyContent, 0666)
			if err != nil {
				panic(err)
			}
			fmt.Fprintf(conn, "HTTP/1.1 201 Created\r\n\r\n")
			return
		} else {

			fileContent, err := os.ReadFile(os.Args[2] + fileName)
			if err != nil {
				fmt.Println("Error reading request.", err.Error())
				fmt.Fprintf(conn, "HTTP/1.1 404 Not Found\r\n\r\n")
				return
			}
			okFile(conn, fileContent)
			return
		}
	}

	switch request.URL.Path {
	case "/":
		fmt.Fprintf(conn, "HTTP/1.1 200 OK\r\n\r\n")
	case "/user-agent":
		okContent(conn, request.UserAgent())
	default:
		fmt.Fprintf(conn, "HTTP/1.1 404 Not Found\r\n\r\n")
	}

}

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")

	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}
	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)

		}

		go handle(conn)
	}
}
