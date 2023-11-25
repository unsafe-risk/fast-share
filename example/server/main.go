package main

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"log"
	"time"

	fastshare "github.com/unsafe-risk/fast-share"
)

func main() {
	const port = 3857

	server, err := fastshare.NewServer(3857, 64)
	if err != nil {
		panic(err)
	}
	defer server.Close()

	go func() {
		if err := server.Listen(port); err != nil {
			panic(err)
		}
	}()

	data := [1024]byte{}
	ticker := time.NewTicker(2 * time.Second)
	for range ticker.C {
		if _, err := rand.Read(data[:]); err != nil {
			panic(err)
		}

		fmt.Printf("Send %d bytes\n", len(data))
		fmt.Printf("Buffer: %x\n", data[:])
		fmt.Println("----------")

		if err := server.Send(binary.BigEndian.Uint32(data[:4]), data[:]); err != nil {
			log.Printf("fast-share.Send: %v", err)
			continue
		}
	}
}
