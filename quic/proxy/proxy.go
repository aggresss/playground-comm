package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"os"

	"github.com/quic-go/quic-go"
)

const (
	listenAddress  = "127.0.0.1:1930"
	forwardAddress = "127.0.0.1:5059"
)

var (
	nextProtos = []string{"quic-echo-example"}
)

func main() {
	listener, err := net.Listen("tcp", listenAddress)
	if err != nil {
		fmt.Println("failed to create listener, err:", err)
		os.Exit(1)
	}
	defer listener.Close()
	fmt.Printf("listening on %s\n", listener.Addr())
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("failed to accept connection, err:", err)
			continue
		}
		go handleConnection(conn)
	}
}

func handleConnection(tcpConn net.Conn) {
	fmt.Printf("accept new connection, remote: %s\n", tcpConn.RemoteAddr().String())
	defer tcpConn.Close()
	tlsConf := &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         nextProtos,
	}
	quicConfig := &quic.Config{}
	quicConn, err := quic.DialAddr(context.Background(), forwardAddress, tlsConf, quicConfig)
	if err != nil {
		fmt.Println("quic dial failed:", err.Error())
		os.Exit(1)
	}
	defer quicConn.CloseWithError(quic.ApplicationErrorCode(quic.NoError), "")
	quicStream, err := quicConn.OpenStreamSync(context.Background())
	if err != nil {
		fmt.Println("quic open stream failed:", err.Error())
		os.Exit(1)
	}
	errSignal := make(chan error, 1)
	go pipe(quicStream, tcpConn, errSignal)
	go pipe(tcpConn, quicStream, errSignal)
	fmt.Printf("start proxy tcp://%s to quic://%s", tcpConn.RemoteAddr().String(), forwardAddress)
	for range errSignal {
		return
	}
	fmt.Printf("stop proxy tcp://%s to quic://%s", tcpConn.RemoteAddr().String(), forwardAddress)
}

func pipe(src, dst io.ReadWriter, errSignal chan error) {
	buff := make([]byte, 0xFFFF) // 64KiB
	for {
		n, err := src.Read(buff)
		if err != nil {
			errSignal <- err
			return
		}
		_, err = dst.Write(buff[:n])
		if err != nil {
			errSignal <- err
			return
		}
	}
}
