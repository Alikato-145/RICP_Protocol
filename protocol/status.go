package protocol

import "encoding/binary"

func EncodeStatus(seq uint32, code uint16) []byte {
	phase, ok := Phase[code]
	if !ok {
		phase = ""
	}
	pb := []byte(phase)
	buf := make([]byte, 8+len(pb))
	binary.BigEndian.PutUint32(buf[0:4], seq)
	buf[4] = TypeStatus
	binary.BigEndian.PutUint16(buf[5:7], code)
	buf[7] = byte(len(pb))
	copy(buf[8:], pb)
	return buf
}

func DecodeStatus(buf []byte) (uint32, uint16, string) {
	seq := binary.BigEndian.Uint32(buf[0:4])
	code := binary.BigEndian.Uint16(buf[5:7])
	n := int(buf[7])
	phrase := string(buf[8 : 8+uint(n)])
	return seq, code, phrase
}
