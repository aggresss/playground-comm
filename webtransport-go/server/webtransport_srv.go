package main

import (
	"bufio"
	"context"
	"crypto/sha256"
	_ "embed"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/quic-go/quic-go/http3"
	"github.com/quic-go/webtransport-go"

	"github.com/aggresss/playground-comm/utils-go"
)

const (
	ADDR = ":5059"
)

//go:embed index.html
var indexHTML string

func runHTTPServer(certHash [32]byte) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		fmt.Println("handler hit")
		content := strings.ReplaceAll(indexHTML, "%%CERTHASH%%", formatByteSlice(certHash[:]))
		w.Write([]byte(content))
	})
	http.ListenAndServe(":8080", mux)
}

func formatByteSlice(b []byte) string {
	s := strings.ReplaceAll(fmt.Sprintf("%#v", b[:]), "[]byte{", "[")
	s = strings.ReplaceAll(s, "}", "]")
	return s
}

func main() {
	tlsConf, err := utils.GetTLSConf(time.Now(), time.Now().Add(10*24*time.Hour))
	if err != nil {
		log.Fatal(err)
	}
	hash := sha256.Sum256(tlsConf.Certificates[0].Leaf.Raw)

	go runHTTPServer(hash)

	wmux := http.NewServeMux()
	s := webtransport.Server{
		H3: http3.Server{
			TLSConfig: tlsConf,
			Addr:      ADDR,
			Handler:   wmux,
		},
		CheckOrigin: func(r *http.Request) bool { return true },
	}
	defer s.Close()

	wmux.HandleFunc("/echo", func(w http.ResponseWriter, r *http.Request) {
		conn, err := s.Upgrade(w, r)
		if err != nil {
			log.Printf("upgrading failed: %s", err)
			w.WriteHeader(500)
			return
		}

		stream, err := conn.AcceptStream(context.Background())
		if err != nil {
			log.Fatalf("failed to accept directional stream: %v", err)
		}
		defer stream.Close()
		fmt.Printf("accept new stream, remote: %s, streamID: %x\n", conn.RemoteAddr().String(), stream.StreamID())

		reader := bufio.NewReader(stream)
		for {
			bytes, err := reader.ReadBytes(byte('\n'))
			if err != nil {
				if err != io.EOF {
					fmt.Println("failed to read data, err:", err)
				}
				return
			}
			fmt.Printf("request: %s", bytes)

			stream.Write(bytes)
			fmt.Printf("response: %s", bytes)
		}
	})
	fmt.Printf("listening on %s\n", s.H3.Addr)
	if err := s.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
