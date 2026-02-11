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

	if len(lines) != 8 {
		panic("INVALID REQUEST")
	}

	var httpreq HTTPReq

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

	secondLine := strings.Split(lines[1], " ")

	if len(secondLine) != 2 {
		panic("INVALID HEADER")
	}

	if secondLine[0] != "Host:" {
		panic("INVALID HEADER")
	}

	httpreq.host = secondLine[1]

	thirdLine := strings.Split(lines[2], " ")

	if thirdLine[0] != "User-Agent:" {
		panic("INVALID HEADER")
	}

	httpreq.userAgent = thirdLine[1]

	forthLine := strings.Split(lines[3], " ")

	if forthLine[0] != "Accept-Encoding:" {
		panic("INVALID HEADER")
	}

	httpreq.acceptEncoding = forthLine[1]

	log.Printf("\n\n\nParsed HTTP Request - Method: %s\nPath: %s\nVersion: %s\nHost: %s\nUser-Agent: %s\nAccept-Encoding: %s\n\n\n",
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
			httpreq := parseHTTPRequest(string(tmp))
			fmt.Println(httpreq)
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
