package main

import (
	"fmt"
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

	//handshake
	conn.Write(protocol.EncodeHello(100, "token123"))
	log.Println("[Controller] -> HELLO token123")
	buf := make([]byte, 1024)
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	n, err := conn.Read(buf)
	if err != nil || n < 5 || buf[4] != protocol.TypeWelcome {
		log.Fatal("handshake error: not received WELCOME")
	}
	_, startSeq := protocol.DecodeWelcome(buf)
	log.Printf("[Controller] <- WELCOME startSeq=%d\n", startSeq)

	var seq uint32 = startSeq
	go func() {
		buf := make([]byte, 1024)
		for {
			n, err := conn.Read(buf)
			if err != nil || n < 8 {
				continue
			}
			if buf[4] == protocol.TypeStatus {
				_, code, phrase := protocol.DecodeStatus(buf)
				fmt.Printf("[Controller] <- %d %s\n", code, phrase)
			}
		}
	}()
	for i := 0; i < 30; i++ {
		x := uint16(i*2 + 30)
		y := uint16(i*3 + 30)
		packet := protocol.EncodeMove(seq, 0, x, y)
		if i%2 == 0 {
			conn.Write(protocol.EncodeStatus(seq, 200))
		} else if i%5 == 0 {
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
