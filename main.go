package main

import (
	"encoding/json"
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
	params      map[string]string
	headers     map[string]string
}

type HTTPResponse struct {
	httpVersion  string
	statusCode   string
	reasonPhrase string
	headers      map[string]string
	body         string
}

func getParams(line string) map[string]string {
	params := make(map[string]string)
	paramsArr := strings.Split(line, "&")

	for _, param := range paramsArr {
		paramKey := strings.Split(param, "=")[0]
		paramValue := strings.Split(param, "=")[1]
		params[paramKey] = paramValue
	}

	return params
}

func getPathAndParams(line string) (string, map[string]string) {
	var path string
	params := make(map[string]string)

	line = line[1:] // To remove the / in the beginning

	// If it doesn't contain any params, then path = line
	if !strings.Contains(line, "?") {
		path = line
		return path, params
	}

	path = strings.Split(line, "?")[0]
	paramsStr := strings.Split(line, "?")[1]
	params = getParams(paramsStr)

	return path, params
}

func parseHTTPRequest(req string) HTTPRequest {
	lines := strings.Split(req, "\n")

	var httpReq HTTPRequest

	// Example of first line: "GET /api/data HTTP/1.1"
	firstLine := strings.Split(lines[0], " ")

	if len(firstLine) != 3 {
		fmt.Printf("INVALID HEADER %s\n", firstLine)
	}

	method := firstLine[0]
	path, params := getPathAndParams(firstLine[1])

	httpReq.params = params

	httpVersion := firstLine[2]

	httpReq.method = method
	httpReq.path = path
	httpReq.httpVersion = httpVersion

	httpReq.headers = make(map[string]string)

	for i := range len(lines) {
		if i == 0 {
			continue
		}

		if httpReq.method == "POST" && i == len(lines)-1 {
			httpReq.headers["Body"] = lines[i]
			continue
		}

		parts := strings.Split(lines[i], ": ")

		if len(parts) < 2 {
			continue
		}

		httpReq.headers[parts[0]] = parts[1]
	}

	log.Printf(
		"\nParsed HTTP Request:\n"+
			"---\n"+
			"Method: %s\n"+
			"Path: %s\n"+
			"Version: %s\n"+
			"Content-Type: %s\n"+
			"Host: %s\n"+
			"User-Agent: %s\n"+
			"Body: %s\n"+
			"---\n",
		httpReq.method,
		httpReq.path,
		httpReq.httpVersion,
		httpReq.headers["Content-Type"],
		httpReq.headers["Host"],
		httpReq.headers["User-Agent"],
		httpReq.headers["Body"],
	)

	return httpReq
}

func handleHTTPRequest(httpReq HTTPRequest) HTTPResponse {
	method := httpReq.method
	path := httpReq.path
	badRequestBody := `{"error":"Bad Request"}`

	var filePath string
	if path == "" {
		filePath = "index.html"
	} else if !strings.Contains(path, ".") {
		// If the path doesn't have extension, we will treat it as json file.
		filePath = strings.Split(path, ".json")[0] + ".json"
	} else if path == "/" {
		filePath = "index.html"
	} else {
		filePath = path
	}

	if method == "GET" {
		// go get the content in the given path and put it in httpRes.body
		_, err := os.Stat(filePath)
		if os.IsNotExist(err) {
			// handle the response
			notFoundHTML := "<h1>404 Not Found</h1>"
			headers := map[string]string{
				"Content-Type":   "text/html",
				"Content-Length": strconv.Itoa(len(string(notFoundHTML))),
				"Connection":     "keep-alive",
			}

			return HTTPResponse{
				httpVersion:  "HTTP/1.1",
				statusCode:   "404",
				reasonPhrase: "Not Found",
				headers:      headers,
				body:         string(notFoundHTML),
			}
		}

		data, err := os.ReadFile(filePath)
		if err != nil {
			fmt.Println("Error: ", err)
		}

		body := string(data)

		if len(httpReq.params) > 0 {
			var result map[string][]map[string]any
			json.Unmarshal(data, &result)

			if _, pageExist := httpReq.params["page"]; pageExist {
				idStr := httpReq.params["page"]
				id, err := strconv.Atoi(idStr)
				if err != nil {
					fmt.Println("Error: ", err)
				}

				// TODO: Send an "Invalid Request" reponse if the id is bigger than the length of the array
				entry := result["pages"][id-1]

				newBody := "{"

				for key, value := range entry {
					switch v := value.(type) {
					case string:
						newBody += fmt.Sprintf(`"%s": "%s"`, key, v)
					case float64:
						// NOTE: All JSON numbers in go become float64
						newBody += fmt.Sprintf(`"%s": "%s",`, key, strconv.Itoa(int(v)))
					default:
						fmt.Printf("unknown type: %T\n", v)
					}
				}

				newBody = newBody[:len(newBody)-1] + "}\n"

				fmt.Println(newBody)
				body = newBody
			}
		}

		return HTTPResponse{
			httpVersion:  "HTTP/1.1",
			statusCode:   "200",
			reasonPhrase: "OK",
			headers: map[string]string{
				"Content-Type":   "application/json",
				"Content-Length": strconv.Itoa(len(body)),
				"Connection":     "keep-alive",
			},
			body: body,
		}
	}

	if method == "POST" {
		// TODO: Better file handling
		file, err := os.Create(filePath)
		if err != nil {
			fmt.Println("Error: ", err)
		}
		defer file.Close()

		var result map[string]string
		body := strings.NewReplacer("\x00", "").Replace(httpReq.headers["Body"])

		if strings.Contains(httpReq.headers["Content-Type"], "x-www-form-urlencoded") {
			result = getParams(body)
		} else {
			err = json.Unmarshal([]byte(body), &result)
			if err != nil {
				fmt.Println(err)
			}
		}

		encoder := json.NewEncoder(file)
		err = encoder.Encode(result)
		if err != nil {
			fmt.Println(err)
		}

		return HTTPResponse{
			httpVersion:  "HTTP/1.1",
			statusCode:   "200",
			reasonPhrase: "OK",
			headers: map[string]string{
				"Content-Length": "0",
				"Connection":     "keep-alive",
			},
		}
	}

	// Return Bad Request
	return HTTPResponse{
		httpVersion:  "HTTP/1.1",
		statusCode:   "400",
		reasonPhrase: "Bad Request",
		headers: map[string]string{
			"Content-Type":   "application/json",
			"Content-Length": strconv.Itoa(len(badRequestBody)),
			"Connection":     "keep-alive",
		},
		body: badRequestBody,
	}
}

// HTTP/1.1 200 OK\r\nContent-Type: text/html\r\nContent-Length: 18\r\nConnection: keep-alive\r\n\r\nHello Vietnaaaaam\r\n\r\n

func getRawHTTPResponse(httpRes HTTPResponse) string {
	rawResponse := fmt.Sprintf("%s %s %s\r\n", httpRes.httpVersion, httpRes.statusCode, httpRes.reasonPhrase)

	for key, value := range httpRes.headers {
		rawResponse += fmt.Sprintf("%s: %s\r\n", key, value)
	}

	rawResponse += fmt.Sprintf("\r\n%s", httpRes.body)

	return rawResponse
}

func handleTCPConnection(connection net.Conn) {
	fmt.Printf("Serving %s\n", connection.RemoteAddr().String())
	tmp := make([]byte, 4096)
	defer connection.Close()
	for {
		_, err := connection.Read(tmp)
		req := string(tmp)
		if strings.Contains(req, "HTTP") {
			httpReq := parseHTTPRequest(string(tmp))
			httpRes := handleHTTPRequest(httpReq)
			httpResRaw := getRawHTTPResponse(httpRes)
			response := fmt.Sprintf(httpResRaw)
			connection.Write([]byte(response))
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
	addr := "localhost:6969"
	fmt.Printf("Listening on %s\n", addr)

	l, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}
	defer l.Close()

	for {
		connection, err := l.Accept()
		if err != nil {
			fmt.Println(err)
			return
		}
		go handleTCPConnection(connection)
	}
}
