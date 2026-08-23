package main

import (
	"log"
	"net"
	"ricp/protocol"
)

func main() {
	conn, err := net.Dial("udp", "127.0.0.1:9200")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	log.Println("Controller has started and connected to the port 9200")
	var seq uint32 = 1000
	for i := 0; i < 20; i++ {
		x := uint16(i*2 + 30)
		y := uint16(i*3 + 30)
		packet := protocol.EncodeMove(seq, 0, x, y)
		_, err := conn.Write(packet)
		if err != nil {
			log.Fatal(err)
		}
		seq++
	}
	log.Println("ส่งครบ 20 datagram")
}
