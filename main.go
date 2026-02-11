package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"strings"
)

type HTTPReq struct {
	method         string
	path           string
	httpVersion    string
	host           string
	userAgent      string
	acceptEncoding string
}

func parseHTTPRequest(req string) HTTPReq {
	lines := strings.Split(req, "\n")

	var httpreq HTTPReq

	// Example of first line: "GET /api/data HTTP/1.1"
	firstLine := strings.Split(lines[0], " ")

	if len(firstLine) != 3 {
		panic("INVALID HEADER")
	}

	method := firstLine[0]
	path := firstLine[1]
	httpVersion := firstLine[2]

	if method != "GET" && method != "POST" {
		panic("INVALID METHOD: " + method)
	}

	httpreq.method = method
	httpreq.path = path
	httpreq.httpVersion = httpVersion

	for i := range len(lines) {
		if i == 0 {
			continue
		}

		parts := strings.Split(lines[i], ": ")

		if parts[0] == "User-Agent" {
			httpreq.userAgent = parts[1]
		}

		if parts[0] == "Host" {
			httpreq.host = parts[1]
		}

		if parts[0] == "Accept-Encoding:" {
			httpreq.acceptEncoding = parts[1]
		}
	}

	log.Printf("\n\n\nParsed HTTP Request:\n---\nMethod: %s\nPath: %s\nVersion: %s\nHost: %s\nUser-Agent: %s\nAccept-Encoding: %s\n---\n",
		httpreq.method, httpreq.path, httpreq.httpVersion, httpreq.host, httpreq.userAgent, httpreq.acceptEncoding)

	return httpreq
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

func handleTCPConnection(c net.Conn) {
	fmt.Printf("Serving %s\n", c.RemoteAddr().String())
	packet := make([]byte, 4096)
	tmp := make([]byte, 4096)
	defer c.Close()
	for {
		_, err := c.Read(tmp)
		req := string(tmp)
		if strings.Contains(req, "HTTP") {
			parseHTTPRequest(string(tmp))
			response := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/html\r\nContent-Length: 13\r\nConnection: keep-alive\r\n\r\nHello, World!\r\n\r\n")
			c.Write([]byte(response))
		}

		if err != nil {
			if err != io.EOF {
				fmt.Println("read error:", err)
			}
			break
		}
		packet = append(packet, tmp...)
	}
}
