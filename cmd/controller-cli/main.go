// RICP Controller (interactive) — รับคำสั่งจาก terminal ทีละบรรทัด
// เหมาะสำหรับ demo/อัดคลิป: พิมพ์สดทีละเคสให้ดู
//
// รัน viewer ก่อน แล้ว:  go run ./cmd/controller-cli
//
// คำสั่ง:
//
//	hello [token]        ทำ handshake (default token123)
//	move <x> <y>         ส่ง MOVE
//	click <x> <y>        ส่ง CLICK (กดซ้าย) — reliable มี ACK
//	key <ตัวอักษร>        ส่ง KEY (เช่น key A) — reliable มี ACK
//	scroll <dx> <dy>     ส่ง SCROLL
//	bad                  ส่ง datagram พัง (2 ไบต์) → viewer ตอบ 400
//	drop                 ข้ามการส่ง 1 datagram แต่ seq เดินต่อ → viewer เห็น GAP
//	bye                  ปิด session
//	help                 แสดงคำสั่ง
//	quit                 ออกจากโปรแกรม
package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"ricp/protocol"
)

var seq uint32 = 1 // seq ก่อน handshake (จะถูกตั้งใหม่จาก WELCOME)

// อ่าน reply (WELCOME/ACK/STATUS/STATS) มาแสดง
func listenReplies(conn net.Conn) {
	buf := make([]byte, 1024)
	for {
		n, err := conn.Read(buf)
		if err != nil || n < 5 {
			continue
		}
		switch buf[4] {
		case protocol.TypeWelcome:
			_, start := protocol.DecodeWelcome(buf)
			seq = start
			fmt.Printf("   <- WELCOME start-seq=%d\n", start)
		case protocol.TypeAck:
			_, ackSeq := protocol.DecodeAck(buf)
			fmt.Printf("   <- ACK seq=%d\n", ackSeq)
		case protocol.TypeStatus:
			_, code, phrase := protocol.DecodeStatus(buf)
			fmt.Printf("   <- STATUS %d %s\n", code, phrase)
		case protocol.TypeStats:
			_, r, l, ro := protocol.DecodeStats(buf)
			fmt.Printf("   <- STATS recv=%d lost=%d reorder=%d\n", r, l, ro)
		}
	}
}

func help() {
	fmt.Println(`คำสั่ง:
  hello [token]     handshake (default: token123)
  move <x> <y>      ส่ง MOVE
  click <x> <y>     ส่ง CLICK (reliable + ACK)
  key <char>        ส่ง KEY เช่น key A
  scroll <dx> <dy>  ส่ง SCROLL
  bad               ส่ง datagram พัง -> 400
  drop              ข้าม 1 datagram (seq เดินต่อ) -> GAP
  bye               ปิด session
  help / quit`)
}
func atoi(s string) int {
	n, _ := strconv.Atoi(s)
	return n
}

func main() {
	conn, err := net.Dial("udp", "127.0.0.1:9200")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	go listenReplies(conn)

	fmt.Println("=== RICP Controller (interactive) — พิมพ์ help ดูคำสั่ง ===")
	sc := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("> ")
		if !sc.Scan() {
			break
		}
		fields := strings.Fields(sc.Text())
		if len(fields) == 0 {
			continue
		}
		cmd := strings.ToLower(fields[0])

		switch cmd {
		case "hello":
			token := "token123"
			if len(fields) > 1 {
				token = fields[1]
			}
			conn.Write(protocol.EncodeHello(0, token))
			seq = 1500
			fmt.Printf("-> HELLO token=%q\n", token)

		case "move":
			if len(fields) < 3 {
				fmt.Println("ใช้: move <x> <y>")
				continue
			}
			conn.Write(protocol.EncodeMove(seq, 0, uint16(atoi(fields[1])), uint16(atoi(fields[2]))))
			fmt.Printf("-> MOVE seq=%d (%s,%s)\n", seq, fields[1], fields[2])
			seq++

		case "click":
			if len(fields) < 3 {
				fmt.Println("ใช้: click <x> <y>")
				continue
			}
			conn.Write(protocol.EncodeClick(seq, 1, 1, 0, uint16(atoi(fields[1])), uint16(atoi(fields[2]))))
			fmt.Printf("-> CLICK seq=%d (%s,%s)\n", seq, fields[1], fields[2])
			seq++

		case "key":
			if len(fields) < 2 {
				fmt.Println("ใช้: key <char>")
				continue
			}
			ch := uint32(fields[1][0]) // ตัวอักษรแรกเป็น keysym (ASCII)
			conn.Write(protocol.EncodeKey(seq, ch, 1, 0))
			fmt.Printf("-> KEY seq=%d key=%c\n", seq, rune(ch))
			seq++

		case "scroll":
			if len(fields) < 3 {
				fmt.Println("ใช้: scroll <dx> <dy>")
				continue
			}
			conn.Write(protocol.EncodeScroll(seq, 0, int8(atoi(fields[1])), int8(atoi(fields[2]))))
			fmt.Printf("-> SCROLL seq=%d (%s,%s)\n", seq, fields[1], fields[2])
			seq++

		case "bad":
			conn.Write([]byte{0xFF, 0xFF}) // สั้นเกิน header -> viewer ตอบ 400
			fmt.Println("-> BAD datagram (2 bytes)")

		case "drop":
			fmt.Printf("-> DROP seq=%d (ไม่ส่ง แต่ seq เดินต่อ)\n", seq)
			seq++ // ข้ามการส่ง แต่ seq เดิน → viewer เห็น GAP

		case "bye":
			conn.Write(protocol.EncodeBye(seq))
			fmt.Printf("-> BYE seq=%d\n", seq)
			seq++

		case "help":
			help()

		case "quit", "exit":
			fmt.Println("bye")
			time.Sleep(100 * time.Millisecond) // ให้ reply ที่ค้างมาถึง
			return

		default:
			fmt.Println("ไม่รู้จักคำสั่ง:", cmd, "(พิมพ์ help)")
		}
	}
}
