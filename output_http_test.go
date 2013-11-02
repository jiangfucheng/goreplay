package gor

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"sync"
	"testing"
)

func startHTTP(cb func(*http.Request)) net.Listener {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cb(r)
	})

	listener, _ := net.Listen("tcp", ":0")

	go http.Serve(listener, handler)

	return listener
}

func TestHTTPOutput(t *testing.T) {
	wg := new(sync.WaitGroup)
	quit := make(chan int)

	input := NewTestInput()

	headers := HTTPHeaders{HTTPHeader{"User-Agent", "Gor"}}

	listener := startHTTP(func(req *http.Request) {
		if req.Header.Get("User-Agent") != "Gor" {
			t.Error("Wrong header")
		}

		wg.Done()
	})

	output := NewHTTPOutput(listener.Addr().String(), headers, "")

	Plugins.Inputs = []io.Reader{input}
	Plugins.Outputs = []io.Writer{output}

	go Start(quit)

	for i := 0; i < 100; i++ {
		wg.Add(2)
		input.EmitGET()
		input.EmitPOST()
	}

	wg.Wait()

	close(quit)
}

func BenchmarkHTTPOutput(b *testing.B) {
	wg := new(sync.WaitGroup)
	quit := make(chan int)

	input := NewTestInput()

	headers := HTTPHeaders{HTTPHeader{"User-Agent", "Gor"}}

	listener := startHTTP(func(req *http.Request) {
		go wg.Done()
	})

	output := NewHTTPOutput(listener.Addr().String(), headers, "")

	Plugins.Inputs = []io.Reader{input}
	Plugins.Outputs = []io.Writer{output}

	go Start(quit)

	fmt.Println(b)

	for i := 0; i < b.N; i++ {
		wg.Add(1)
		input.EmitPOST()
	}

	wg.Wait()

	close(quit)
}
