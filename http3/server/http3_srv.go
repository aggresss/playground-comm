package main

import (
	// "fmt"
	// "io"

	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/http3"
)

const (
	ADDR = ":5059"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("handle request from %s\n", r.RemoteAddr)
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			body = []byte(fmt.Sprintf("error reading request body: %s", err))
		}
		w.Write([]byte(body))
	})

	quicConfig := &quic.Config{}

	server := http3.Server{
		Addr:       ADDR,
		QuicConfig: quicConfig,
	}
	fmt.Printf("listening on %s\n", server.Addr)
	if err := server.ListenAndServeTLS("server.crt", "server.key"); err != nil {
		log.Fatal(err)
	}
}
