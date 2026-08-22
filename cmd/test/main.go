package main

import (
	"fmt"

	"ricp/protocol" // module "ricp" (จาก go.mod) + โฟลเดอร์ "protocol"
)

func main() {
	pkt := protocol.EncodeMove(1000, protocol.TypeMove, 300, 250)
	fmt.Printf("MOVE datagram: % X\n", pkt)

	seq, x, y, mask := protocol.DecodeMove(pkt)
	fmt.Printf("decode: seq=%d x=%d y=%d mask=%08b\n", seq, x, y, mask)

	// เข้าถึง constant ก็ได้
	fmt.Printf("TypeKey = 0x%02X\n", protocol.TypeKey)
}
