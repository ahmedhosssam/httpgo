package main

import (
    "fmt"
    "io"
    "log"
    "net"
    "strings"
)

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
            fmt.Println(string(tmp))
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
