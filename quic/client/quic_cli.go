package main

import (
	"bufio"
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

const message = "foobar\n"

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

	quicConfig := &quic.Config{}

	conn, err := quic.DialAddr(quicServer.String(), tlsConf, quicConfig)
	if err != nil {
		fmt.Println("Dial failed:", err.Error())
		os.Exit(1)
	}

	stream, err := conn.OpenStreamSync(context.Background())
	if err != nil {
		fmt.Println("OpenStreamSync failed:", err.Error())
		os.Exit(1)
	}

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)

	readerStream := bufio.NewReader(stream)
	readerStdin := bufio.NewReader(os.Stdin)

	for {
		select {
		case s := <-ch:
			stream.Close()
			conn.CloseWithError(quic.ApplicationErrorCode(quic.NoError), "")
			fmt.Println(s)
			os.Exit(0)
		default:
			text, _ := readerStdin.ReadString('\n')
			_, err = stream.Write([]byte(text))
			if err != nil {
				fmt.Println("Write data failed:", err.Error())
				os.Exit(1)
			}
			fmt.Printf("send: %s", text)

			bytes, err := readerStream.ReadBytes(byte('\n'))
			if err != nil {
				if err != io.EOF {
					fmt.Println("failed to read data, err:", err)
				}
				os.Exit(1)
			}
			fmt.Printf("receive: %s", bytes)
		}
	}
}
