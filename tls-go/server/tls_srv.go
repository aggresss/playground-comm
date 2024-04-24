package main

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/aggresss/playground-comm/utils-go"
)

func main() {
	config, err := utils.GetTLSConf(time.Now(), time.Now().Add(10*24*time.Hour))
	if err != nil {
		return
	}

	listener, err := tls.Listen("tcp", ":5059", config)
	if err != nil {
		fmt.Println("failed to create listener, err:", err)
		return
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

func handleConnection(conn net.Conn) {
	fmt.Printf("accept new connection, remote: %s\n", conn.RemoteAddr().String())
	defer conn.Close()
	reader := bufio.NewReader(conn)
	for {
		bytes, err := reader.ReadBytes(byte('\n'))
		if err != nil {
			if err != io.EOF {
				fmt.Println("failed to read data, err:", err)
			}
			return
		}
		fmt.Printf("request: %s", bytes)

		conn.Write(bytes)
		fmt.Printf("response: %s", bytes)
	}
}
