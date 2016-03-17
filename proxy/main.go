package main

import (
	"bufio"
	"log"
	"net"
	"net/http"
)

func main() {
	// Listen tcp on port 8080
	if ln, err := net.Listen("tcp", ":8080"); err == nil {
		// Loop: accept connections one by one
		for {
			// Accept tcp connection
			if conn, err := ln.Accept(); err == nil {
				// Read from connection
				reader := bufio.NewReader(conn)
				// Read HTTP
				if req, err := http.ReadRequest(reader); err == nil {
					// Connect to backend
					if be, err := net.Dial("tcp", "127.0.0.1:8081"); err == nil {
						// Read from the backend connection
						be_reader := bufio.NewReader(be)

						// Write to the backend
						if err := req.Write(be); err == nil {
							// Read response from the backend
							if resp, err := http.ReadResponse(be_reader, req); err == nil {
								// Send the response to the client
								resp.Close = true
								if err := resp.Write(conn); err == nil {
									log.Printf("%s: %d", req.URL.Path, resp.StatusCode)
								}

								conn.Close()
								// Loop: accept next connection
							}
						}
					}
				}
			}
		}
	}
}
