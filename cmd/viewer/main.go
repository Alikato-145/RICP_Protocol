package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"ricp/protocol" // module "ricp" (จาก go.mod) + โฟลเดอร์ "protocol"
)

type viewerState struct {
	authenticated bool
	lastSeq       uint32
	received      uint32
	lost          uint32
	reordered     uint32
	first         bool
}

func setupUDP() (*net.UDPConn, error) {
	addr, err := net.ResolveUDPAddr("udp", ":9200")
	if err != nil {
		return nil, err
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func (st *viewerState) trackSeq(seq uint32) (isDrop bool) {
	if !st.first && seq > st.lastSeq+1 {
		fmt.Printf("gap detected: lastSeq=%d, seq=%d\n", st.lastSeq, seq)
		st.lost += seq - st.lastSeq - 1

	}
	if !st.first && seq <= st.lastSeq {
		fmt.Printf("reorder/dup: seq=%d\n", seq)
		st.reordered++
		return true
	}
	st.lastSeq = seq
	st.first = false
	return false
}

func handleDatagram(conn *net.UDPConn, buf []byte, n int, src *net.UDPAddr, st *viewerState) {
	type_message := buf[4]
	if !st.authenticated && type_message != protocol.TypeHello {
		fmt.Println("not authenticated")
		conn.WriteToUDP(protocol.EncodeStatus(0, 401), src)
		return
	}
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
		if token == "token123" {
			startSeq := uint32(1500)
			st.authenticated = true
			st.lastSeq = startSeq - 1
			conn.WriteToUDP(protocol.EncodeWelcome(0, startSeq), src)
			log.Printf("[Viewer] -> WELCOME start-seq=%d", startSeq)
		} else {
			conn.WriteToUDP(protocol.EncodeStatus(0, 401), src)
		}
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
		conn.WriteToUDP(protocol.EncodeStatus(0, 400), src)
	}
	st.received++
	fmt.Printf("total: received=%d, lost=%d, reordered=%d\n", st.received, st.lost, st.reordered)
}

func main() {
	conn, err := setupUDP()
	st := viewerState{first: true}
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	fmt.Println("Viewr listening UDP at port : 9200")
	buf := make([]byte, 1024)
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
		if st.trackSeq(seq) {
			continue
		}
		fmt.Printf("seq: %d ", seq)
		handleDatagram(conn, buf[:n], n, src, &st)
	}

}
