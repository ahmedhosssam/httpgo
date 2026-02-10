package main

import (
    "fmt"
    "io"
    "log"
    "net"
    "strings"
)

type HTTPReq struct {
    method string
    path string
    http_version string
    host string
    user_agent string
    accept_encoding string
}

func parse_http_request(req string) HTTPReq {
    lines := strings.Split(req, "\n")

    if len(lines) != 8 {
        panic("INVALID REQUEST")
    }

    var httpreq HTTPReq

    first_line := strings.Split(lines[0], " ")

    if len(first_line) != 3 {
        panic("INVALID HEADER")
    }

    method := first_line[0]
    path := first_line[1]
    http_version := first_line[2]

    if method != "GET" && method != "POST" {
        panic("INVALID METHOD: " + method)
    }

    httpreq.method = method
    httpreq.path = path
    httpreq.http_version = http_version

    second_line := strings.Split(lines[1], " ")

    if len(second_line) != 2 {
        panic("INVALID HEADER")
    }

    if second_line[0] != "Host:" {
        panic("INVALID HEADER")
    }

    httpreq.host = second_line[1]

    third_line := strings.Split(lines[2], " ")

    if third_line[0] != "User-Agent:" {
        panic("INVALID HEADER")
    }

    httpreq.user_agent = third_line[1]

    forth_line := strings.Split(lines[3], " ")

    if forth_line[0] != "Accept-Encoding:" {
        panic("INVALID HEADER")
    }

    httpreq.accept_encoding = forth_line[1]

    log.Printf("Parsed HTTP Request - Method: %s\nPath: %s\nVersion: %s\nHost: %s\nUser-Agent: %s\nAccept-Encoding: %s",
        httpreq.method, httpreq.path, httpreq.http_version, httpreq.host, httpreq.user_agent, httpreq.accept_encoding)

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
        go handle_tcp_connection(c)
    }
}

func handle_tcp_connection(c net.Conn) {
    fmt.Printf("Serving %s\n", c.RemoteAddr().String())
    packet := make([]byte, 4096)
    tmp := make([]byte, 4096)
    defer c.Close()
    for {
        _, err := c.Read(tmp)
        req := string(tmp)
        if strings.Contains(req, "HTTP") {
            httpreq := parse_http_request(string(tmp))
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
