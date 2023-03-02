package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/quic-go/quic-go"
)

const (
	HOST      = "localhost"
	PORT      = "5059"
	TYPE      = "tcp"
	NEXTPROTO = "quic-echo-example"
)

const message = "foobar"

func main() {
	quicServer, err := net.ResolveTCPAddr(TYPE, HOST+":"+PORT)
	if err != nil {
		fmt.Println("ResolveTCPAddr failed:", err.Error())
		os.Exit(1)
	}

	tlsConf := &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         []string{NEXTPROTO},
	}

	quicConfig := quic.Config{}

	conn, err := quic.DialAddr(quicServer.String(), tlsConf, &quicConfig)
	if err != nil {
		fmt.Println("Dial failed:", err.Error())
		os.Exit(1)
	}

	stream, err := conn.OpenStreamSync(context.Background())
	if err != nil {
		return
	}

	fmt.Printf("Client: Sending '%s'\n", message)
	_, err = stream.Write([]byte(message))
	if err != nil {
		fmt.Println("stream write failed:", err.Error())
		os.Exit(1)
	}

	buf := make([]byte, len(message))
	_, err = io.ReadFull(stream, buf)
	if err != nil {
		fmt.Println("stream read failed:", err.Error())
		os.Exit(1)
	}
	fmt.Printf("Client: Got '%s'\n", buf)

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	select {
	case s := <-ch:
		fmt.Println(s.String())
	}
}
