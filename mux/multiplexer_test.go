package mux

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"testing"
)

func serveHttp1(l net.Listener) {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		a := r.URL.Query().Get("a")
		fmt.Printf("a: %v \n", a)
		io.WriteString(w, "hello HTTP1")
	})

	if err := http.Serve(l, http.DefaultServeMux); err != ErrListenerClosed {
		panic(err)
	}
}

func TestServe(t *testing.T) {
	ol, _ := net.Listen("tcp", ":8080")
	mux := NewMux(ol, 1024)
	nl := mux.RegisterNewProtocolListener(HTTP1())
	//nl := mux.RegisterNewProtocolListener(func(_ io.Reader) bool {
	//	return true
	//})

	go serveHttp1(nl)

	//go func() {
	//	time.Sleep(1 * time.Second)
	//	mux.Close()
	//}()

	if err := mux.Serve(); err != ErrServerClosed {
		panic(err)
	}
}
