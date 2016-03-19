package main

import (
	"bufio"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"sync"
)

var requestBytes map[string]int64
var requestLock sync.Mutex

func init() {
	requestBytes = make(map[string]int64)
}

func updateStats(req *http.Request, resp *http.Response) int64 {
	requestLock.Lock()
	defer requestLock.Unlock()

	bytes := requestBytes[req.URL.Path] + resp.ContentLength
	requestBytes[req.URL.Path] = bytes
	return bytes
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)

	for {
		req, err := http.ReadRequest(reader)
		if err != nil {
			if err != io.EOF {
				log.Printf("Failed to read request: %s", err)
			}
			return
		}
		if backendConn, err := net.Dial("tcp", "127.0.0.1:8081"); err == nil {
			backend_reader := bufio.NewReader(backendConn)
			if err := req.Write(backendConn); err == nil {
				if resp, err := http.ReadResponse(backend_reader, req); err == nil {
					bytes := updateStats(req, resp)
					resp.Header.Set("X-Bytes", strconv.FormatInt(bytes, 10))
					resp.Close = true
					if err := resp.Write(conn); err == nil {
						log.Printf("%s: %d, %d bytes", req.URL.Path, resp.StatusCode, requestBytes[req.URL.Path])
					}
				}
			}
		}
	}
}

func main() {
	ln, err := net.Listen("tcp", ":"+os.Args[1])
	if err != nil {
		log.Fatalf("Failed to listen: %s", err)
	}
	for {
		if conn, err := ln.Accept(); err == nil {
			log.Print("Received connection: %v", conn)
			go handleConnection(conn)
		}
	}
}
