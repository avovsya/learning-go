package main

import (
	"log"
	"net/rpc"
	"strconv"
)

type Empty struct {
}

type Stats struct {
	RequestBytes map[string]int64
}

func getLongestUrl(m map[string]int64) int {
	var result int

	for k := range m {
		if len(k) > result {
			result = len(k)
		}
	}

	return result
}

func printStats(stats Stats) {
	maxSpaces := strconv.Itoa(getLongestUrl(stats.RequestBytes) + 2)

	log.Println("Proxy Stats:")
	log.Printf("| URL%"+maxSpaces+"s| Bytes |", " ")
	log.Printf("|    %"+maxSpaces+"s|       |", " ")

	for k, v := range stats.RequestBytes {
		log.Printf("| %-"+maxSpaces+"s   |%6d |", k, v)
	}
}

func main() {
	client, err := rpc.DialHTTP("tcp", "127.0.0.1:8079")

	if err != nil {
		log.Fatalf("Failed to dial: %s", err)
	}

	var reply Stats
	err = client.Call("RpcServer.GetStats", &Empty{}, &reply)
	if err != nil {
		log.Fatalf("Failed to GetStats: %s", err)
	}
	printStats(reply)
}
