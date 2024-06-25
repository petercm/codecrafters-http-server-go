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

type HttpResponse struct {
	out     io.Writer
	status  string
	headers map[string]string
	content []byte
}

func okString(resp HttpResponse, content string) {
	resp.status = "200 OK"
	resp.headers["Content-Type"] = "text/plain"
	resp.headers["Content-Length"] = fmt.Sprint(len(content))
	resp.content = []byte(content)
	respond(resp)
}

func okFile(resp HttpResponse, content []byte) {
	resp.status = "200 OK"
	resp.headers["Content-Type"] = "application/octet-stream"
	resp.headers["Content-Length"] = fmt.Sprint(len(content))
	resp.content = content
	respond(resp)
}

func respond(resp HttpResponse) {
	fmt.Fprintf(resp.out, "HTTP/1.1 %s\r\n", resp.status)
	if len(resp.headers) > 0 {
		for header, value := range resp.headers {
			fmt.Fprintf(resp.out, "%s: %s\r\n", header, value)
		}
	} else {
		fmt.Fprintf(resp.out, "\r\n")
	}
	fmt.Fprintf(resp.out, "\r\n") // end headers
	fmt.Fprintf(resp.out, "%s\r\n", resp.content)
}

func handle(conn net.Conn) {
	defer conn.Close()
	request, err := http.ReadRequest(bufio.NewReader(conn))
	if err != nil {
		fmt.Println("Error reading request.", err.Error())
	}
	fmt.Printf("Request: %s %s \n", request.Method, request.URL)
	resp := HttpResponse{out: conn, headers: make(map[string]string)}
	acceptEncoding := request.Header["Accept-Encoding"]
	if len(acceptEncoding) > 0 && acceptEncoding[0] == "gzip" {
		resp.headers["Content-Encoding"] = "gzip"
	}

	toEcho, isEcho := strings.CutPrefix(request.URL.Path, "/echo/")
	if isEcho {
		okString(resp, toEcho)
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
			resp.status = "201 Created"
			respond(resp)
			return
		} else {

			fileContent, err := os.ReadFile(os.Args[2] + fileName)
			if err != nil {
				fmt.Println("Error reading request.", err.Error())
				resp.status = "404 Not Found"
				respond(resp)
				return
			}
			okFile(resp, fileContent)
			return
		}
	}

	switch request.URL.Path {
	case "/":
		resp.status = "200 OK"
		respond(resp)
	case "/user-agent":
		okString(resp, request.UserAgent())
	default:
		resp.status = "404 Not Found"
		respond(resp)
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
