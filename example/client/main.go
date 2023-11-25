package main

import (
	"bytes"
	"fmt"

	fastshare "github.com/unsafe-risk/fast-share"
)

func main() {
	const port = 3857

	client := fastshare.NewClient(port)

	if err := client.Connect(); err != nil {
		panic(err)
	}

	if err := client.Attach(); err != nil {
		panic(err)
	}
	defer client.Close()

	buf := bytes.NewBuffer(nil)
	for {
		buf.Reset()

		name, length, err := client.Receive(buf)
		if err != nil {
			panic(err)
		}

		fmt.Printf("Received %d bytes for %d\n", length, name)
		fmt.Printf("Buffer Length: %d\n", buf.Len())
		fmt.Printf("Buffer: %x\n", buf.Bytes())
		fmt.Println("----------")
	}
}
