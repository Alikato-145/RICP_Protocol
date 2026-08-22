package protocol

import "testing"

func TestMoveRoundtrip(t *testing.T) {
	pkt := EncodeMove(1000, 1, 300, 250)
	if len(pkt) != 10 {
		t.Fatalf("MOVE should be 10 bytes, got %d", len(pkt))
	}

	seq, mask, x, y := DecodeMove(pkt)
	if seq != 1000 || mask != 1 || x != 300 || y != 250 || pkt[4] != TypeMove {
		t.Fatalf("decode mismatch: seq=%d mask=%d x=%d y=%d type=%d", seq, mask, x, y, pkt[4])
	}
}
func TestClickRoundtrip(t *testing.T) {
	pkt := EncodeClick(1299, 1, 1, 0)
	if len(pkt) != 8 {
		t.Fatalf("CLICK should be 8 bytes, got %d", len(pkt))
	}

	seq, mask, down_flag, reserved := DecodeClick(pkt)
	if seq != 1299 || mask != 1 || down_flag != 1 || reserved != 0 || pkt[4] != TypeClick {
		t.Fatalf("decode mismatch: seq=%d mask=%d down_flag=%d reserved=%d type=%d", seq, mask, down_flag, reserved, pkt[4])
	}
}

func TestKeyRoundtrip(t *testing.T) {
	pkt := EncodeKey(1001, 1, 'a', 0)
	if len(pkt) != 11 {
		t.Fatalf("KEY should be 11 bytes, got %d", len(pkt))
	}

	seq, mask, key, reserved := DecodeKey(pkt)
	if seq != 1001 || mask != 1 || key != 'a' || reserved != 0 || pkt[4] != TypeKey {
		t.Fatalf("decode mismatch: seq=%d mask=%d key=%c reserved=%d type=%d", seq, mask, key, reserved, pkt[4])
	}
}

func TestScrollRoundtrip(t *testing.T) {
	pkt := EncodeScroll(1002, 1, 123, 123)
	if len(pkt) != 10 {
		t.Fatalf("SCROLL should be 10 bytes, got %d", len(pkt))
	}

	seq, mask, dx, dy := DecodeScroll(pkt)
	if seq != 1002 || mask != 1 || dx != 123 || dy != 123 || pkt[4] != TypeScroll {
		t.Fatalf("decode mismatch: seq=%d mask=%d dx=%d dy=%d type=%d", seq, mask, dx, dy, pkt[4])
	}
}

func TestHelloRoundtrip(t *testing.T) {
	pkt := EncodeHello(1, "test-token123")
	if len(pkt) != 19 || pkt[5] != 13 {
		t.Fatalf("HELLO should be 19 bytes, got %d (length=%d)", len(pkt), pkt[5])
	}

	seq, token := DecodeHello(pkt)
	if seq != 1 || token != "test-token123" || pkt[4] != TypeHello {
		t.Fatalf("decode mismatch: seq=%d token=%s type=%d", seq, token, pkt[4])
	}
}

func TestWelcomeRoundtrip(t *testing.T) {
	pkt := EncodeWelcome(123, 456)
	if len(pkt) != 9 {
		t.Fatalf("WELCOME should be 9 bytes, got %d", len(pkt))
	}

	seq, start := DecodeWelcome(pkt)
	if seq != 123 || start != 456 || pkt[4] != TypeWelcome {
		t.Fatalf("decode mismatch: seq=%d start=%d type=%d", seq, start, pkt[4])
	}
}

func TestAckRoundtrip(t *testing.T) {
	pkt := EncodeAck(1003, 123)
	if len(pkt) != 9 {
		t.Fatalf("ACK should be 5 bytes, got %d", len(pkt))
	}

	seq, seq_ack := DecodeAck(pkt)
	if seq != 1003 || seq_ack != 123 || pkt[4] != TypeAck {
		t.Fatalf("decode mismatch: seq=%d seq_ack=%d type=%d", seq, seq_ack, pkt[4])
	}
}

func TestStatsRoundtrip(t *testing.T) {
	pkt := EncodeStats(123, 12, 3, 4)
	if len(pkt) != 17 {
		t.Fatalf("STATS should be 17 bytes, got %d", len(pkt))
	}

	seq, received, lost, reordered := DecodeStats(pkt)
	if seq != 123 || received != 12 || lost != 3 || reordered != 4 || pkt[4] != TypeStats {
		t.Fatalf("decode mismatch: seq=%d received=%d lost=%d reordered=%d type=%d", seq, received, lost, reordered, pkt[4])
	}
}

func TestByeRoundtrip(t *testing.T) {
	pkt := EncodeBye(32)
	if len(pkt) != 5 {
		t.Fatalf("BYE should be 5 bytes, got %d", len(pkt))
	}

	seq := DecodeBye(pkt)
	if seq != 32 || pkt[4] != TypeBye {
		t.Fatalf("decode mismatch: seq=%d type=%d", seq, pkt[4])
	}
}
