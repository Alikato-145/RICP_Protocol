package protocol

import (
	"encoding/binary"
)

func EncodeMove(seq uint32, mask uint8, x, y uint16) []byte {
	buf := make([]byte, 10)
	binary.BigEndian.PutUint32(buf[0:4], seq)
	buf[4] = TypeMove 
	binary.BigEndian.PutUint16(buf[5:7], x)
	binary.BigEndian.PutUint16(buf[7:9], y)
	buf[9] = mask
	return buf
}

func DecodeMove(buf []byte) (seq uint32, mask uint8, x, y uint16) {
	seq = binary.BigEndian.Uint32(buf[0:4])
	x = binary.BigEndian.Uint16(buf[5:7])
	y = binary.BigEndian.Uint16(buf[7:9])
	mask = buf[9]
	return seq, mask, x, y
}

func EncodeClick(seq uint32, mask uint8, down_flag uint8, reserved uint8) []byte {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint32(buf[0:4], seq)
	buf[4] = TypeClick
	buf[5] = mask
	buf[6] = down_flag
	buf[7] = reserved
	return buf
}

func DecodeClick(buf []byte) (seq uint32, mask uint8, down_flag uint8, reserved uint8) {
	seq = binary.BigEndian.Uint32(buf[0:4])
	mask = buf[5]
	down_flag = buf[6]
	reserved = buf[7]
	return seq, mask, down_flag, reserved
}

// keySystem follow X11 RFC 6143
func EncodeKey(seq uint32, keysym uint32, down_flag uint8, reserved uint8) []byte {
	buf := make([]byte, 11)
	binary.BigEndian.PutUint32(buf[0:4], seq)
	buf[4] = TypeKey
	binary.BigEndian.PutUint32(buf[5:9], keysym)
	buf[9] = down_flag
	buf[10] = reserved
	return buf
}

func DecodeKey(buf []byte) (seq uint32, keysym uint32, down_flag uint8, reserved uint8) {
	seq = binary.BigEndian.Uint32(buf[0:4])
	keysym = binary.BigEndian.Uint32(buf[5:9])
	down_flag = buf[9]
	reserved = buf[10]
	return seq, keysym, down_flag, reserved
}

func EncodeScroll(seq uint32, mask uint8, x, y int8) []byte {
	buf := make([]byte, 10)
	binary.BigEndian.PutUint32(buf[0:4], seq)
	buf[4] = TypeScroll
	buf[5] = mask
	buf[6] = byte(x)
	buf[7] = byte(y)
	return buf
}

func DecodeScroll(buf []byte) (seq uint32, mask uint8, x, y int8) {
	seq = binary.BigEndian.Uint32(buf[0:4])
	mask = buf[5]
	x = int8(buf[6])
	y = int8(buf[7])
	return seq, mask, x, y
}
func EncodeHello(seq uint32, token string) []byte {
	tb := []byte(token)
	buf := make([]byte, 6+len(tb))
	binary.BigEndian.PutUint32(buf[0:4], seq)
	buf[4] = TypeHello
	buf[5] = uint8(len(tb)) // คำนวณเอง ไม่รับ length มา
	copy(buf[6:], tb)
	return buf
}
func DecodeHello(buf []byte) (seq uint32, token string) {
	seq = binary.BigEndian.Uint32(buf[0:4])
	n := int(buf[5])
	token = string(buf[6 : 6+n])
	return seq, token
}
func EncodeWelcome(seq uint32, start uint32) []byte {
	buf := make([]byte, 9)
	binary.BigEndian.PutUint32(buf[0:4], seq)
	buf[4] = TypeWelcome
	binary.BigEndian.PutUint32(buf[5:9], start)
	return buf
}
func DecodeWelcome(buf []byte) (seq uint32, start uint32) {
	seq = binary.BigEndian.Uint32(buf[0:4])
	start = binary.BigEndian.Uint32(buf[5:9])
	return seq, start
}

func EncodeAck(seq uint32, ack_seq uint32) []byte {
	buf := make([]byte, 9)
	binary.BigEndian.PutUint32(buf[0:4], seq)
	buf[4] = TypeAck
	binary.BigEndian.PutUint32(buf[5:9], ack_seq)
	return buf
}
func DecodeAck(buf []byte) (seq uint32, ack_seq uint32) {
	seq = binary.BigEndian.Uint32(buf[0:4])
	ack_seq = binary.BigEndian.Uint32(buf[5:9])
	return seq, ack_seq
}

func EncodeStats(seq uint32, received uint32, lost uint32, reordered uint32) []byte {
	buf := make([]byte, 17)
	binary.BigEndian.PutUint32(buf[0:4], seq)
	buf[4] = TypeStats
	binary.BigEndian.PutUint32(buf[5:9], received)
	binary.BigEndian.PutUint32(buf[9:13], lost)
	binary.BigEndian.PutUint32(buf[13:17], reordered)
	return buf
}
func DecodeStats(buf []byte) (seq uint32, received uint32, lost uint32, reordered uint32) {
	seq = binary.BigEndian.Uint32(buf[0:4])
	received = binary.BigEndian.Uint32(buf[5:9])
	lost = binary.BigEndian.Uint32(buf[9:13])
	reordered = binary.BigEndian.Uint32(buf[13:17])
	return seq, received, lost, reordered
}

func EncodeBye(seq uint32) []byte {
	buf := make([]byte, 5)
	binary.BigEndian.PutUint32(buf[0:4], seq)
	buf[4] = TypeBye
	return buf
}
func DecodeBye(buf []byte) uint32 {
	return binary.BigEndian.Uint32(buf[0:4])
}
