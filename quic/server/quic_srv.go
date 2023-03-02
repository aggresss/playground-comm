package main

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/quic-go/quic-go"

	"github.com/aggresss/playground-comm/utils"
)

const (
	ADDR      = ":5059"
	NEXTPROTO = "quic-echo-example"
)

func main() {
	tlsConfig, err := utils.GetTLSConf(time.Now(), time.Now().Add(10*24*time.Hour))
	if err != nil {
		return
	}
	tlsConfig.NextProtos = []string{NEXTPROTO}

	quicConfig := quic.Config{}

	listener, err := quic.ListenAddr(ADDR, tlsConfig, &quicConfig)
	if err != nil {
		fmt.Println("failed to create listener, err:", err)
		return
	}

	defer listener.Close()

	fmt.Printf("listening on %s\n", listener.Addr())
	for {
		conn, err := listener.Accept(context.Background())
		if err != nil {
			fmt.Println("failed to accept connection, err:", err)
			continue
		}

		go handleConnection(conn)
	}
}

func handleConnection(conn quic.Connection) {
	stream, err := conn.AcceptStream(context.Background())
	if err != nil {
		panic(err)
	}
	// Echo through the loggingWriter
	_, err = io.Copy(loggingWriter{stream}, stream)
}

// A wrapper for io.Writer that also logs the message.
type loggingWriter struct{ io.Writer }

func (w loggingWriter) Write(b []byte) (int, error) {
	fmt.Printf("Server: Got '%s'\n", string(b))
	return w.Writer.Write(b)
}
