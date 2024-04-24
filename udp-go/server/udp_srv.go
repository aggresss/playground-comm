package main

import (
	"fmt"
	"log"
	"net"
	"time"
)

func main() {
	// listen to incoming udp packets
	udpServer, err := net.ListenPacket("udp", ":1053")
	if err != nil {
		log.Fatal(err)
	}
	defer udpServer.Close()
	fmt.Printf("listening on %s\n", udpServer.LocalAddr().String())

	for {
		buf := make([]byte, 1024)
		_, addr, err := udpServer.ReadFrom(buf)
		if err != nil {
			continue
		}
		go response(udpServer, addr, buf)
	}

}

func response(udpServer net.PacketConn, addr net.Addr, buf []byte) {
	time := time.Now().Format(time.ANSIC)
	responseStr := fmt.Sprintf("time received: %v.  -> Your message: %v!", time, string(buf))

	udpServer.WriteTo([]byte(responseStr), addr)
}
