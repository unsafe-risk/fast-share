package main

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"net"
	"runtime/debug"
	"sync"
	"time"

	fastshare "github.com/unsafe-risk/fast-share"
)

const Repeat = 100000
const bufSize = 16384

func main() {
	s := time.Now()
	BenchmarkFastshare()
	e := time.Since(s)
	fmt.Printf("fastshare: %v\n", e)

	debug.FreeOSMemory()

	s = time.Now()
	BenchmarkLoopbackTcp()
	e = time.Since(s)
	fmt.Printf("loopback tcp: %v\n", e)
}

func BenchmarkFastshare() {
	wg := sync.WaitGroup{}
	wg.Add(2)

	go func() {
		server, err := fastshare.NewServer(3850, 32678)
		if err != nil {
			panic(err)
		}
		defer wg.Done()
		defer server.Close()

		go func() {
			if err := server.Listen(3857); err != nil {
				panic(err)
			}
		}()

		time.Sleep(1 * time.Second)

		buf := [bufSize]byte{}
		for i := 0; i < Repeat; i++ {
			if _, err := rand.Read(buf[:]); err != nil {
				panic(err)
			}

			if err := server.Send(0, buf[:]); err != nil {
				panic(err)
			}
		}
	}()

	go func() {
		client := fastshare.NewClient("localhost:3857")

		if err := client.Connect(); err != nil {
			panic(err)
		}

		if err := client.Attach(); err != nil {
			panic(err)
		}
		defer wg.Done()
		defer client.Close()

		buf := bytes.NewBuffer(nil)
		for i := 0; i < Repeat; i++ {
			buf.Reset()
			if _, _, err := client.Receive(buf); err != nil {
				panic(err)
			}
		}
	}()

	wg.Wait()
}

func BenchmarkLoopbackTcp() {
	wg := sync.WaitGroup{}
	wg.Add(2)

	go func() {
		lis, err := net.Listen("tcp", ":8080")
		if err != nil {
			panic(err)
		}

		defer wg.Done()
		defer lis.Close()

		time.Sleep(1 * time.Second)

		conn, err := lis.Accept()
		if err != nil {
			panic(err)
		}

		buf := [bufSize]byte{}
		for i := 0; i < Repeat; i++ {
			if _, err := rand.Read(buf[:]); err != nil {
				panic(err)
			}

			if _, err := conn.Write(buf[:]); err != nil {
				panic(err)
			}
		}

		conn.Close()
	}()

	go func() {
		conn, err := net.Dial("tcp", "localhost:8080")
		if err != nil {
			panic(err)
		}

		defer wg.Done()
		defer conn.Close()

		buf := [bufSize]byte{}
		for i := 0; i < Repeat; i++ {
			if _, err := conn.Read(buf[:]); err != nil {
				panic(err)
			}
		}
	}()

	wg.Wait()
}
