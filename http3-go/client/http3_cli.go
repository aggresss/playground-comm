package main

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/http3"
)

const (
	HOST = "localhost"
	PORT = "5059"
	TYPE = "tcp"
)

func main() {
	tlsConf := &tls.Config{
		InsecureSkipVerify: true,
	}

	quicConfig := &quic.Config{}

	t := &http3.RoundTripper{
		TLSClientConfig: tlsConf,
		QuicConfig:      quicConfig,
	}

	defer t.Close()

	client := http.Client{Transport: t, Timeout: 15 * time.Second}

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("https://%s", HOST+":"+PORT), bytes.NewBuffer([]byte("Hello, World!")))
	if err != nil {
		log.Fatalf("unable to create http request due to error %s", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		switch e := err.(type) {
		case *url.Error:
			log.Fatalf("url.Error received on http request: %s", e)
		default:
			log.Fatalf("Unexpected error received: %s", err)
		}
	}

	body, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		log.Fatalf("unexpected error reading response body: %s", err)
	}

	fmt.Printf("\nResponse from server: \n\tHTTP status: %s\n\tBody: %s\n", resp.Status, body)
}
