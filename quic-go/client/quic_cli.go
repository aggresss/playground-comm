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

	"github.com/quic-go/quic-go"
)

const (
	HOST      = "localhost"
	PORT      = "5059"
	TYPE      = "tcp"
	NEXTPROTO = "sample"
	KEYLOG    = "key.log"
)

func main() {
	keyLog, err := os.OpenFile(KEYLOG, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not open %s\n", KEYLOG)
		os.Exit(1)
	}
	defer keyLog.Close()

	quicServer, err := net.ResolveTCPAddr(TYPE, HOST+":"+PORT)
	if err != nil {
		fmt.Println("ResolveTCPAddr failed:", err.Error())
		os.Exit(1)
	}

	tlsConf := &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         []string{NEXTPROTO},
		KeyLogWriter:       keyLog,
	}

	quicConfig := &quic.Config{
		EnableDatagrams:       true,
		Disable1RTTEncryption: true,
	}

	conn, err := quic.DialAddr(context.Background(), quicServer.String(), tlsConf, quicConfig)
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
	signal.Notify(ch, os.Interrupt)

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
