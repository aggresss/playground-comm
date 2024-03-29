package main

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"os"
)

const (
	HOST   = "localhost"
	PORT   = "5059"
	TYPE   = "tcp"
	KEYLOG = "key.log"
)

func main() {
	tlsServer, err := net.ResolveTCPAddr(TYPE, HOST+":"+PORT)
	if err != nil {
		fmt.Println("ResolveTCPAddr failed:", err.Error())
		os.Exit(1)
	}

	keyLog, err := os.OpenFile(KEYLOG, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not open %s\n", KEYLOG)
		os.Exit(1)
	}
	defer keyLog.Close()

	conf := &tls.Config{
		InsecureSkipVerify: true,
		KeyLogWriter:       keyLog,
	}

	conn, err := tls.Dial(TYPE, tlsServer.String(), conf)
	if err != nil {
		fmt.Println("Dial failed:", err.Error())
		os.Exit(1)
	}
	defer conn.Close()

	readerConn := bufio.NewReader(conn)
	readerStdin := bufio.NewReader(os.Stdin)

	for {
		text, _ := readerStdin.ReadString('\n')
		_, err = conn.Write([]byte(text))
		if err != nil {
			fmt.Println("Write data failed:", err.Error())
			os.Exit(1)
		}
		fmt.Printf("send: %s", text)

		bytes, err := readerConn.ReadBytes(byte('\n'))
		if err != nil {
			if err != io.EOF {
				fmt.Println("failed to read data, err:", err)
			}
			os.Exit(1)
		}
		fmt.Printf("receive: %s", bytes)
	}
}
