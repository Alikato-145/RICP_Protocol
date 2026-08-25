package main

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"ricp/protocol"
	"sync"
	"time"
)

// ---------- event: input for pending ----------
type event struct {
	kind byte // TypeMove / TypeClick / TypeKey
	x, y uint16
	mask uint8
	down uint8
	key  uint32
}

// ---------- pending: rest event before flush (coalescing) ----------
type pending struct {
	lastMove  *event  // MOVE
	reliableQ []event // CLICK/KEY
}

// --------- reliableTracker: tracks unacked datagrams ---------
type reliableTracker struct {
	mu      sync.Mutex
	unacked map[uint32][]byte    //seq -> datagram
	sentAt  map[uint32]time.Time //seq -> sent time
}

func (p *pending) add(e event) {
	if e.kind == protocol.TypeMove {
		p.lastMove = &e
	} else {
		p.reliableQ = append(p.reliableQ, e) // add to queue
	}
}

// flush: encode + send all package
func (p *pending) flush(conn net.Conn, seq *uint32, rt *reliableTracker) {
	movesCoalesced := 0
	if p.lastMove != nil {
		movesCoalesced = 1
	}

	// send reliable (CLICK/KEY)
	for _, e := range p.reliableQ {
		switch e.kind {
		case protocol.TypeClick:
			data := protocol.EncodeClick(*seq, e.mask, e.down, 0)
			conn.Write(data)
			rt.mu.Lock()
			rt.unacked[*seq] = data
			rt.sentAt[*seq] = time.Now()
			rt.mu.Unlock()
		case protocol.TypeKey:
			data := protocol.EncodeKey(*seq, e.key, e.down, 0)
			conn.Write(data)
			rt.mu.Lock()
			rt.unacked[*seq] = data
			rt.sentAt[*seq] = time.Now()
			rt.mu.Unlock()
		}
		*seq++
	}

	// send MOVE received from lastMove
	if p.lastMove != nil {
		m := p.lastMove
		//conn.Write(protocol.EncodeMove(*seq, m.mask, m.x, m.y))
		// test drop
		if rand.Intn(100) < 20 {
			log.Printf("[DROP] seq=%d", *seq)
		} else {
			conn.Write(protocol.EncodeMove(*seq, m.mask, m.x, m.y))
		}
		*seq++
	}

	if movesCoalesced > 0 || len(p.reliableQ) > 0 {
		log.Printf("[FLUSH] moves=%d keys/clicks=%d", movesCoalesced, len(p.reliableQ))
	}

	// reset
	p.lastMove = nil
	p.reliableQ = nil
}

// ---------- handshake ----------
func doHandshake(conn net.Conn) (uint32, error) {
	conn.Write(protocol.EncodeHello(0, "token123"))
	log.Println("[Controller] -> HELLO token123")
	buf := make([]byte, 1024)
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	n, err := conn.Read(buf)
	if err != nil || n < 5 || buf[4] != protocol.TypeWelcome {
		return 0, fmt.Errorf("handshake error: not received WELCOME")
	}
	_, startSeq := protocol.DecodeWelcome(buf)
	conn.SetReadDeadline(time.Time{}) // ปลด deadline
	log.Printf("[Controller] <- WELCOME startSeq=%d", startSeq)
	return startSeq, nil
}

// ---------- goroutine read reply (STATUS/ACK/STATS) ----------
func listenReplies(conn net.Conn, rt *reliableTracker) {
	buf := make([]byte, 1024)
	for {
		n, err := conn.Read(buf)
		if err != nil || n < 5 {
			continue
		}
		switch buf[4] {
		case protocol.TypeStatus:
			_, code, phrase := protocol.DecodeStatus(buf)
			fmt.Printf("[Controller] <- %d %s\n", code, phrase)
		case protocol.TypeAck:
			_, ackSeq := protocol.DecodeAck(buf)
			fmt.Printf("[Controller] <- ACK seq=%d\n", ackSeq)
			rt.mu.Lock()
			delete(rt.unacked, ackSeq)
			delete(rt.sentAt, ackSeq)
			rt.mu.Unlock()
		case protocol.TypeStats:
			_, r, l, ro := protocol.DecodeStats(buf)
			fmt.Printf("[Controller] <- STATS received=%d lost=%d reordered=%d\n", r, l, ro)
		}
	}
}
func (rt *reliableTracker) resendLoop(conn net.Conn) {
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()
	for range ticker.C {
		rt.mu.Lock()
		for seq, data := range rt.unacked {
			if time.Since(rt.sentAt[seq]) > 100*time.Millisecond {
				_, err := conn.Write(data)
				if err != nil {
					log.Println(err)
				}
			}
		}
		rt.mu.Unlock()
	}
}

func main() {
	conn, err := net.Dial("udp", "127.0.0.1:9200")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	// 1) handshake
	startSeq, err := doHandshake(conn)
	if err != nil {
		log.Fatal(err)
	}
	seq := startSeq

	// 2) goroutine read replies and resend unacked datagrams
	rt := reliableTracker{
		unacked: make(map[uint32][]byte),
		sentAt:  make(map[uint32]time.Time),
	}
	go rt.resendLoop(conn)
	go listenReplies(conn, &rt)

	// 3) producer: send events
	events := make(chan event, 100)
	go func() {
		for i := 0; i < 100; i++ {
			events <- event{kind: protocol.TypeMove, x: uint16(i), y: uint16(i * 2)}
			if i%3 == 0 {
				events <- event{kind: protocol.TypeClick, mask: 1, down: 1}
			}
			if i%20 == 0 {
				events <- event{kind: protocol.TypeKey, key: 'A', down: 1}
			}
			time.Sleep(8 * time.Millisecond)
		}
		close(events)
	}()

	// 4) sender: flush event every 50ms → make coalescing
	var p pending
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case e, ok := <-events:
			if !ok {
				p.flush(conn, &seq, &rt) // last flush after pipe closed
				log.Println("ส่งครบแล้ว")
				return
			}
			p.add(e)
		case <-ticker.C:
			p.flush(conn, &seq, &rt)
		}
	}
}
