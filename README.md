# httpgo

A simple HTTP server built from scratch in Go using raw TCP sockets.

## Usage

```sh
go run main.go
```

The server starts on `localhost:6969`.

Right now the most useful use case for httpgo is to host a web page or SPA (single page application) by serving `index.html` and linking css and js files to it.

## Examples

```sh
# Get a file
curl http://localhost:6969/index.html

# Get JSON with pagination
curl http://localhost:6969/search?page=1

# Post JSON data
curl -X POST http://localhost:6969/data -H "Content-Type: application/json" -d '{"name":"ahmed"}'
```
