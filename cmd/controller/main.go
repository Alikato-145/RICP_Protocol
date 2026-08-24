package main

import (
	"log"
	"math/rand"
	"net"
	"ricp/protocol"
	"time"
)

func main() {
	conn, err := net.Dial("udp", "127.0.0.1:9200")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	log.Println("Controller has started and connected to the port 9200")
	var seq uint32 = 1000
	for i := 0; i < 30; i++ {
		x := uint16(i*2 + 30)
		y := uint16(i*3 + 30)
		packet := protocol.EncodeMove(seq, 0, x, y)

		if i%5 == 0 {
			seq++
			conn.Write(protocol.EncodeClick(seq, 1, 1, 0)) // กดซ้าย
		} else if i == 10 {
			seq++
			conn.Write(protocol.EncodeKey(seq, 'A', 1, 0))
		} else if rand.Intn(100) < 20 { // 20% โอกาส
			log.Printf("[DROP] seq=%d (จำลอง packet loss)", seq)
		} else {
			conn.Write(packet)
		}
		time.Sleep(100 * time.Millisecond)
		seq++
	}
	log.Println("ส่งครบ 20 datagram")
}
