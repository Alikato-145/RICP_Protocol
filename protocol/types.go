package protocol

// รหัสชนิดข้อความ (message type) — เป็น byte (0-255) ไม่ใช่ string
const (
	TypeMove    = 0x01 // x (U16), y (U16), button-mask (U8)
	TypeClick   = 0x02 // button-mask (U8), down-flag (U8), reserved (U8)
	TypeKey     = 0x03 // down-flag (U8), reserved (U8), keysym (U32)
	TypeScroll  = 0x04 // delta-x (S8), delta-y (S8), reserved (U8)
	TypeHello   = 0x10 // เริ่ม session
	TypeWelcome = 0x11 // start-seq
	TypeAck     = 0x12 // ack-seq (U32) => ยืนยันรับ CLICK/KEY
	TypeStats   = 0x13 // received (U32), lost (U32), reordered (U32)
	TypeStatus  = 0x20 // code (U16), phraseLen (U8), phrase
	TypeBye     = 0x1F // ปิด session
)
