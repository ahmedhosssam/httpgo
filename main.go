package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
)

type HTTPRequest struct {
	method      string
	path        string
	httpVersion string
	headers     map[string]string
}

type HTTPResponse struct {
	httpVersion  string
	statusCode   string
	reasonPhrase string
	headers      map[string]string
	body         string
}

func parseHTTPRequest(req string) HTTPRequest {
	lines := strings.Split(req, "\n")

	var httpReq HTTPRequest

	// Example of first line: "GET /api/data HTTP/1.1"
	firstLine := strings.Split(lines[0], " ")

	if len(firstLine) != 3 {
		panic("INVALID HEADER")
	}

	method := firstLine[0]
	path := firstLine[1]
	httpVersion := firstLine[2]

	httpReq.method = method
	httpReq.path = path[1:] // To remove the / in the beginning
	httpReq.httpVersion = httpVersion

	httpReq.headers = make(map[string]string)

	for i := range len(lines) {
		if i == 0 {
			continue
		}

		parts := strings.Split(lines[i], ": ")

		if len(parts) < 2 {
			continue
		}

		httpReq.headers[parts[0]] = parts[1]
	}

	log.Printf(
		"\n\n\nParsed HTTP Request:\n"+
			"---\n"+
			"Method: %s\n"+
			"Path: %s\n"+
			"Version: %s\n"+
			"Host: %s\n"+
			"User-Agent: %s\n"+
			"Accept-Encoding: %s\n"+
			"---\n",
		httpReq.method,
		httpReq.path,
		httpReq.httpVersion,
		httpReq.headers["Host"],
		httpReq.headers["User-Agent"],
		httpReq.headers["Accept-Encoding"],
	)

	return httpReq
}

func getHTTPResponse(httpReq HTTPRequest) HTTPResponse {
	var httpRes HTTPResponse
	httpRes.headers = make(map[string]string)

	method := httpReq.method
	path := httpReq.path

	if method != "GET" && method != "POST" {
		httpRes.httpVersion = "HTTP/1.1"
		httpRes.statusCode = "400 "
		httpRes.reasonPhrase = "Bad Request"
		httpRes.headers["Content-Type"] = "text/html"
		// TODO: Body should be like:
		// {
		//   "error": "Bad Request"
		// }
	}

	if method == "GET" {
		// go get the content in the given path and put it in httpRes.body
		data, err := os.ReadFile(path)
		if err != nil {
			fmt.Println("Error: ", err)
		}

		httpRes.body = string(data)
		httpRes.httpVersion = "HTTP/1.1"
		httpRes.statusCode = "200"
		httpRes.reasonPhrase = "OK"
		httpRes.headers["Content-Type"] = "text/html"
		httpRes.headers["Content-Length"] = strconv.Itoa(len(string(data)))
		httpRes.headers["Connection"] = "keep-alive"
	}

	return httpRes
}

// HTTP/1.1 200 OK\r\nContent-Type: text/html\r\nContent-Length: 18\r\nConnection: keep-alive\r\n\r\nHello Vietnaaaaam\r\n\r\n

func getRawHTTPResponse(httpRes HTTPResponse) string {
	var rawResponse string

	rawResponse += httpRes.httpVersion
	rawResponse += " "
	rawResponse += httpRes.statusCode
	rawResponse += " "
	rawResponse += httpRes.reasonPhrase
	rawResponse += "\r\n"

	for key, value := range httpRes.headers {
		rawResponse += key + ": " + value + "\r\n"
	}

	rawResponse += "\r\n"
	rawResponse += httpRes.body

	fmt.Println(rawResponse)
	return rawResponse
}

func handleTCPConnection(c net.Conn) {
	fmt.Printf("Serving %s\n", c.RemoteAddr().String())
	tmp := make([]byte, 4096)
	defer c.Close()
	for {
		_, err := c.Read(tmp)
		req := string(tmp)
		if strings.Contains(req, "HTTP") {
			httpReq := parseHTTPRequest(string(tmp))
			httpRes := getHTTPResponse(httpReq)
			httpResRaw := getRawHTTPResponse(httpRes)
			fmt.Println(httpResRaw)
			response := fmt.Sprintf(httpResRaw)
			c.Write([]byte(response))
		}

		if err != nil {
			if err != io.EOF {
				fmt.Println("read error:", err)
			}
			break
		}
	}
}

func main() {
	PORT := "localhost:6969"
	l, err := net.Listen("tcp4", PORT)
	if err != nil {
		log.Fatal(err)
	}
	defer l.Close()

	for {
		c, err := l.Accept()
		if err != nil {
			fmt.Println(err)
			return
		}
		go handleTCPConnection(c)
	}
}
