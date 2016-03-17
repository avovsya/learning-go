package main

import (
	"bufio"
	"io"
	"log"
	"net"
	"net/http"
	"os"
)

func handleConnection(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)

	// for {
	req, err := http.ReadRequest(reader)
	if err != nil {
		if err != io.EOF {
			log.Printf("Failed to read request: %s", err)
		}
		return
	}
	log.Println("Dialing")
	if backendConn, err := net.Dial("tcp", "127.0.0.1:8081"); err == nil {
		backend_reader := bufio.NewReader(backendConn)
		log.Println("Writing to target server")
		if err := req.Write(backendConn); err == nil {
			log.Println("Reading response")
			if resp, err := http.ReadResponse(backend_reader, req); err == nil {
				resp.Close = true
				log.Println("Writing response")
				if err := resp.Write(conn); err == nil {
					log.Printf("%s: %d", req.URL.Path, resp.StatusCode)
				}
			}
		}
	}
	// }
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
