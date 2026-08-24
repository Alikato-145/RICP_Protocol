package main

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"ricp/protocol"
	"time"
)

func doHandshake(conn net.Conn) (uint32, error) {
	conn.Write(protocol.EncodeHello(100, "token123"))
	log.Println("[Controller] -> HELLO token123")
	buf := make([]byte, 1024)
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	n, err := conn.Read(buf)
	if err != nil || n < 5 || buf[4] != protocol.TypeWelcome {
		return 0, fmt.Errorf("handshake error: not received WELCOME")
	}
	_, startSeq := protocol.DecodeWelcome(buf)
	log.Printf("[Controller] <- WELCOME startSeq=%d\n", startSeq)
	return startSeq, nil
}
func listenStatus(conn net.Conn) {
	buf := make([]byte, 1024)
	for {
		n, err := conn.Read(buf)
		if err != nil || n < 8 {
			continue
		}
		switch buf[4] {
		case protocol.TypeStatus:
			_, code, phrase := protocol.DecodeStatus(buf)
			fmt.Printf("[Controller] <- %d %s\n", code, phrase)
		case protocol.TypeAck:
			_, seq := protocol.DecodeAck(buf)
			fmt.Printf("[Controller] <- ACK seq=%d\n", seq)
		case protocol.TypeStats:
			_, received, lost, reordered := protocol.DecodeStats(buf)
			fmt.Printf("[Controller] <- STATS received=%d lost=%d reordered=%d\n", received, lost, reordered)
		}
	}
}

func main() {
	conn, err := net.Dial("udp", "127.0.0.1:9200")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	startSeq, err := doHandshake(conn)
	if err != nil {
		log.Fatal(err)
	}
	var seq uint32 = startSeq
	go listenStatus(conn)
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
