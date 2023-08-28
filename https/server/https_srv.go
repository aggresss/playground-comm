package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/aggresss/playground-comm/utils"
)

const (
	ADDR = ":5059"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("handle request from %s\n", r.RemoteAddr)
		body, err := io.ReadAll(r.Body)
		if err != nil {
			body = []byte(fmt.Sprintf("error reading request body: %s", err))
		}
		w.Write([]byte(body))
	})

	tlsConfig, err := utils.GetTLSConf(time.Now(), time.Now().Add(10*24*time.Hour))
	if err != nil {
		return
	}

	server := &http.Server{
		Addr:      ADDR,
		TLSConfig: tlsConfig,
	}

	fmt.Printf("listening on %s\n", server.Addr)
	if err := server.ListenAndServeTLS("", ""); err != nil {
		log.Fatal(err)
	}
}
