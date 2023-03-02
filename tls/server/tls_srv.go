package srv

import (
	"bufio"
	"crypto/tls"
	"log"
	"net"
	"time"

	"github.com/aggresss/playground-comm/utils"
)

func main() {
	log.SetFlags(log.Lshortfile)

	config, err := utils.GetTLSConf(time.Now(), time.Now().Add(10*24*time.Hour))
	if err != nil {
		return
	}

	ln, err := tls.Listen("tcp", ":443", config)
	if err != nil {
		log.Println(err)
		return
	}
	defer ln.Close()

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	r := bufio.NewReader(conn)
	for {
		msg, err := r.ReadString('\n')
		if err != nil {
			log.Println(err)
			return
		}

		println(msg)

		n, err := conn.Write([]byte("world\n"))
		if err != nil {
			log.Println(n, err)
			return
		}
	}
}
