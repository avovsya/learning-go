package main

/*
*	TODO
* [ ] Keep backend connection pull warn
* [ ] Get rid of old backend connections in the pull
* [ ] Prettify code(separate to multiple files)
* [ ] Write benchmark
* [ ] Compare to https://github.com/zorkian/lca2015/tree/master/final
 */

import (
	"bufio"
	"io"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"strconv"
	"sync"
	"time"
)

type Backend struct {
	net.Conn
	Reader *bufio.Reader
	Writer *bufio.Writer
}

var backendQueue chan *Backend
var requestBytes map[string]int64
var requestLock sync.Mutex

func init() {
	requestBytes = make(map[string]int64)
	backendQueue = make(chan *Backend, 10)
}

func queueBackend(be *Backend) {
	select {
	case backendQueue <- be:
	case <-time.After(1 * time.Second):
		be.Close()
	}
}

func getBackend() (*Backend, error) {
	select {
	case be := <-backendQueue:
		return be, nil
	case <-time.After(100 * time.Millisecond):
		be, err := net.Dial("tcp", "127.0.0.1:8081")
		if err != nil {
			return nil, err
		}

		return &Backend{
			Conn:   be,
			Reader: bufio.NewReader(be),
			Writer: bufio.NewWriter(be),
		}, nil
	}
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

		be, err := getBackend()
		if err != nil {
			return
		}

		if err := req.Write(be.Writer); err == nil {
			be.Writer.Flush()
			if resp, err := http.ReadResponse(be.Reader, req); err == nil {
				respWriter := bufio.NewWriter(conn)
				bytes := updateStats(req, resp)
				resp.Header.Set("X-Bytes", strconv.FormatInt(bytes, 10))
				resp.Close = true
				if err := resp.Write(respWriter); err == nil {
					respWriter.Flush()
					log.Printf("%s: %d, %d bytes", req.URL.Path, resp.StatusCode, bytes)
				}

				if resp.Close {
					return
				}
			}
		}

		go queueBackend(be)
	}
}

func main() {
	rpc.Register(&RpcServer{})
	rpc.HandleHTTP()
	l, err := net.Listen("tcp", ":8079")
	if err != nil {
		log.Fatalf("Failed to listen for RPC: %s", err)
	}
	go http.Serve(l, nil)

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
