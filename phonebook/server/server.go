package main

import (
	"bufio"
	"fmt"
	// "io"
	"log"
	"net"
)

func readLine(r *bufio.Reader) (string, error) {
	var buffer []byte
	buffer = make([]byte, 10)

	for {

		line, isPrefix, err := r.ReadLine()

		if err != nil {
			return "", err
		}

		buffer = append(buffer, line...)

		if !isPrefix {
			return string(buffer[:]), nil
		}

	}
}

func prompt(r *bufio.Reader, w *bufio.Writer) {
	for {
		w.WriteString("\nHello, how are you?\n")
		w.Flush()
		res, _ := readLine(r)
		w.WriteString(fmt.Sprintf("Your response: %s", res))
		w.Flush()
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)

	log.Printf("Incoming connection from: %s", conn.RemoteAddr())

	prompt(reader, writer)
}

func main() {
	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatal(err)
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		handleConnection(conn)
	}
}
