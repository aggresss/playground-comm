package main

import (
	"bufio"
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

	quicConfig := &quic.Config{
		EnableDatagrams: true,
	}

	listener, err := quic.ListenAddr(ADDR, tlsConfig, quicConfig)
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

		handleConnection(conn)
	}
}

func handleConnection(conn quic.Connection) {
	fmt.Printf("accept new connection, remote: %s\n", conn.RemoteAddr().String())

	go func() {
		defer conn.CloseWithError(quic.ApplicationErrorCode(quic.NoError), "")

		for {
			message, err := conn.ReceiveMessage(context.Background())
			if err != nil {
				fmt.Println("failed to receive message, err:", err)
				break
			}
			fmt.Printf("receive message: %s", string(message))
		}
	}()

	go func() {
		for {
			stream, err := conn.AcceptStream(context.Background())
			if err != nil {
				fmt.Println("failed to accept stream, err:", err)
				return
			}
			fmt.Printf("accept new stream, remote: %s, streamID: %x\n", conn.RemoteAddr().String(), stream.StreamID())

			go handleStream(stream)
		}
	}()
}

func handleStream(stream quic.Stream) {
	defer stream.Close()
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
}
