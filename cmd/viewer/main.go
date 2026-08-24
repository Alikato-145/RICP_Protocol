package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"ricp/protocol" // module "ricp" (จาก go.mod) + โฟลเดอร์ "protocol"
)

func main() {
	var received, lost, reordered uint32
	addr, err := net.ResolveUDPAddr("udp", ":9200")
	if err != nil {
		log.Fatal(err)
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	fmt.Println("Viewr listening UDP at port : 9200")

	buf := make([]byte, 1024)
	var lastSeq uint32
	first := true
	for {
		n, src, err := conn.ReadFromUDP(buf)
		if err != nil {
			fmt.Printf("read error: %s\n", err)
			continue
		}
		if n < 5 {
			conn.WriteToUDP(protocol.EncodeStatus(0, 400), src)
			continue
		}
		seq := binary.BigEndian.Uint32(buf[:4])
		if !first && seq > lastSeq+1 {
			fmt.Printf("gap detected: lastSeq=%d, seq=%d\n", lastSeq, seq)
			lost += seq - lastSeq - 1

		}
		if !first && seq <= lastSeq {
			fmt.Printf("reorder/dup: seq=%d\n", seq)
			reordered++
			continue
		}
		lastSeq = seq
		first = false
		type_message := buf[4]
		fmt.Printf("seq: %d ", seq)
		switch type_message {
		case protocol.TypeMove:
			_, mask, x, y := protocol.DecodeMove(buf)
			fmt.Printf("move: mask=%d, x=%d, y=%d\n", mask, x, y)
		case protocol.TypeClick:
			_, mask, down_flag, reserved := protocol.DecodeClick(buf)
			fmt.Printf("click: mask=%d, down=%d, reserved=%d\n", mask, down_flag, reserved)
		case protocol.TypeKey:
			_, key, down_flag, reserved := protocol.DecodeKey(buf)
			fmt.Printf("key: key=%d, down=%d, reserved=%d\n", key, down_flag, reserved)
		case protocol.TypeScroll:
			_, mask, x, y := protocol.DecodeScroll(buf)
			fmt.Printf("scroll: mask=%d, x=%d, y=%d\n", mask, x, y)
		case protocol.TypeHello:
			_, token := protocol.DecodeHello(buf)
			fmt.Printf("hello: token=%s\n", token)
		case protocol.TypeWelcome:
			_, start := protocol.DecodeWelcome(buf)
			fmt.Printf("welcome: start=%d\n", start)
		case protocol.TypeAck:
			_, ack_seq := protocol.DecodeAck(buf)
			fmt.Printf("ack: seq=%d\n", ack_seq)
		case protocol.TypeStats:
			_, r, l, ro := protocol.DecodeStats(buf)
			fmt.Printf("stats: received=%d, lost=%d, reordered=%d\n", r, l, ro)
		case protocol.TypeBye:
			s := protocol.DecodeBye(buf)
			fmt.Printf("bye: %d\n", s)
		case protocol.TypeStatus:
			_, code, phrase := protocol.DecodeStatus(buf)
			fmt.Printf("status: code=%d, phrase=%s\n", code, phrase)
		default:
			fmt.Printf("unknown type: %d\n", type_message)
			conn.WriteToUDP(protocol.EncodeStatus(seq, 400), src)
		}
		received++
		fmt.Printf("total: received=%d, lost=%d, reordered=%d\n", received, lost, reordered)
	}

}
