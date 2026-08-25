package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"ricp/protocol"
	"sync"
	"time"
)

type viewerState struct {
	mu            sync.Mutex
	authenticated bool
	lastSeq       uint32
	received      uint32
	lost          uint32
	reordered     uint32
	first         bool
	clientAddr    *net.UDPAddr
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
		missing := seq - st.lastSeq - 1
		st.mu.Lock()
		st.lost += missing
		fmt.Printf("  !! GAP      lost %d packet (expected %d, got %d)\n", missing, st.lastSeq+1, seq)
		st.mu.Unlock()
	}
	if !st.first && seq <= st.lastSeq {
		st.mu.Lock()
		st.reordered++
		fmt.Printf("  !! REORDER  dropped late/dup seq=%d\n", seq)
		st.mu.Unlock()
		return true
	}
	st.mu.Lock()
	st.lastSeq = seq
	st.first = false
	st.mu.Unlock()
	return false
}

func handleDatagram(conn *net.UDPConn, buf []byte, seq uint32, st *viewerState) {
	type_message := buf[4]
	if !st.authenticated && type_message != protocol.TypeHello {
		fmt.Printf("[seq %-5d] DENIED   not authenticated -> 401\n", seq)
		conn.WriteToUDP(protocol.EncodeStatus(0, 401), st.clientAddr)
		return
	}
	switch type_message {
	case protocol.TypeMove:
		_, mask, x, y := protocol.DecodeMove(buf)
		fmt.Printf("[seq %-5d] MOVE     x=%-4d y=%-4d mask=%d\n", seq, x, y, mask)
	case protocol.TypeClick:
		_, mask, down, _, x, y := protocol.DecodeClick(buf)
		fmt.Printf("[seq %-5d] CLICK    mask=%d down=%d x=%d y=%d\n", seq, mask, down, x, y)
		//ACK after receiving action
		conn.WriteToUDP(protocol.EncodeAck(0, seq), st.clientAddr)

	case protocol.TypeKey:
		_, key, down, _ := protocol.DecodeKey(buf)
		fmt.Printf("[seq %-5d] KEY      key=%c down=%d\n", seq, rune(key), down)
		//ACK after receiving action
		conn.WriteToUDP(protocol.EncodeAck(0, seq), st.clientAddr)
	case protocol.TypeScroll:
		_, mask, x, y := protocol.DecodeScroll(buf)
		fmt.Printf("[seq %-5d] SCROLL   dx=%d dy=%d mask=%d\n", seq, x, y, mask)
	case protocol.TypeHello:
		_, token := protocol.DecodeHello(buf)
		if token == "token123" {
			startSeq := uint32(1500)
			st.mu.Lock()
			st.authenticated = true
			st.lastSeq = startSeq - 1
			st.mu.Unlock()
			conn.WriteToUDP(protocol.EncodeWelcome(0, startSeq), st.clientAddr)
			fmt.Printf("[seq %-5d] HELLO    token=%q -> WELCOME start-seq=%d\n", seq, token, startSeq)
		} else {
			conn.WriteToUDP(protocol.EncodeStatus(0, 401), st.clientAddr)
			fmt.Printf("[seq %-5d] HELLO    token=%q -> 401 UNAUTHORIZED\n", seq, token)
		}
		return // handshake ไม่นับเป็น input
	case protocol.TypeBye:
		fmt.Printf("[seq %-5d] BYE      session closed\n", seq)
		st.mu.Lock()
		st.authenticated = false
		st.first = true
		st.received, st.lost, st.reordered = 0, 0, 0
		st.mu.Unlock()
		return
	case protocol.TypeStatus:
		_, code, phrase := protocol.DecodeStatus(buf)
		fmt.Printf("[seq %-5d] STATUS   %d %s\n", seq, code, phrase)
	default:
		fmt.Printf("[seq %-5d] UNKNOWN  type=0x%02X -> 400\n", seq, type_message)
		conn.WriteToUDP(protocol.EncodeStatus(0, 400), st.clientAddr)
		return
	}
	st.mu.Lock()
	st.received++
	r, l, ro := st.received, st.lost, st.reordered
	st.mu.Unlock()
	fmt.Printf("            stats: recv=%d lost=%d reorder=%d\n", r, l, ro)
}
func (st *viewerState) reportStats(conn *net.UDPConn) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		st.mu.Lock()
		r, l, ro := st.received, st.lost, st.reordered
		addr := st.clientAddr
		auth := st.authenticated
		st.mu.Unlock()
		if addr != nil && auth {
			conn.WriteToUDP(protocol.EncodeStats(0, r, l, ro), addr)
			fmt.Printf(">> STATS sent: recv=%d lost=%d reorder=%d\n", r, l, ro)
		}
	}
}

func main() {
	conn, err := setupUDP()
	st := viewerState{first: true}
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	fmt.Println("=== RICP Viewer listening on UDP :9200 ===")
	buf := make([]byte, 1024)
	go st.reportStats(conn)
	for {
		n, src, err := conn.ReadFromUDP(buf)
		if err != nil {
			fmt.Printf("read error: %s\n", err)
			continue
		}
		st.mu.Lock()
		st.clientAddr = src
		st.mu.Unlock()
		if err != nil {
			fmt.Printf("read error: %s\n", err)
			continue
		}
		if n < 5 {
			fmt.Printf("  !! BAD       datagram too short (%d bytes) -> 400\n", n)
			conn.WriteToUDP(protocol.EncodeStatus(0, 400), src)
			continue
		}
		seq := binary.BigEndian.Uint32(buf[:4])
		msgType := buf[4]
		st.mu.Lock()
		isOld := !st.first && seq <= st.lastSeq
		st.mu.Unlock()
		if msgType == protocol.TypeHello {
			handleDatagram(conn, buf[:n], seq, &st)
			continue
		}
		if isOld && (msgType == protocol.TypeClick || msgType == protocol.TypeKey) {
			conn.WriteToUDP(protocol.EncodeAck(0, seq), src)
			continue
		}
		if st.trackSeq(seq) {
			continue
		}
		handleDatagram(conn, buf[:n], seq, &st)
	}
}
