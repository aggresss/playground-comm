package main

import (
	"bufio"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/http3"
	"github.com/quic-go/webtransport-go"
)

const (
	HOST = "localhost"
	PORT = "5059"
)

const message = "foobar\n"

func main() {

	tlsConf := &tls.Config{
		InsecureSkipVerify: true,
	}

	quicConfig := &quic.Config{}

	d := webtransport.Dialer{
		RoundTripper: &http3.RoundTripper{
			TLSClientConfig: tlsConf,
			QuicConfig:      quicConfig,
		},
	}

	resp, conn, err := d.Dial(context.Background(), fmt.Sprintf("https://%s:%s/echo", HOST, PORT), nil)
	if err != nil {
		fmt.Println("Dial failed:", err.Error())
		os.Exit(1)
	}
	if resp.StatusCode != 200 {
		fmt.Println("Dial response abnormal")
		os.Exit(1)
	}

	stream, err := conn.OpenStreamSync(context.Background())
	if err != nil {
		fmt.Println("OpenStreamSync failed:", err.Error())
		return
	}

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)

	readerStream := bufio.NewReader(stream)
	readerStdin := bufio.NewReader(os.Stdin)

	for {
		select {
		case s := <-ch:
			stream.Close()
			conn.CloseWithError(0, "")
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
