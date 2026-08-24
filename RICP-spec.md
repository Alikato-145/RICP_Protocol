# RICP/1.0 — Remote Input Control Protocol
**ร่างข้อกำหนดโพรโทคอลชั้นแอปพลิเคชัน (Application-Layer Protocol Specification)**

> โพรโทคอลสำหรับส่งเหตุการณ์อินพุต (เมาส์เคลื่อนที่ / คลิก / กดคีย์) จากเครื่องผู้ควบคุม
> ไปแสดงผลที่เครื่องผู้รับแบบเรียลไทม์ ผ่าน UDP โดยจัดการ packet loss และการสลับลำดับ
> เองที่ชั้นแอปพลิเคชัน

---

## 1. วัตถุประสงค์ของโปรแกรม

**RICP Remote Pointer** เป็นโปรแกรมส่งต่อเหตุการณ์อินพุต (input event relay) แบบ
client–server เครื่อง **Controller** (ผู้ส่ง) ดักเหตุการณ์การใช้เมาส์และคีย์บอร์ดของผู้ใช้
แล้วส่งผ่านเครือข่ายไปยังเครื่อง **Viewer** (ผู้รับ) ซึ่งนำเหตุการณ์เหล่านั้นมา
**แสดงผลแบบจำลอง** (วาดตำแหน่ง pointer, แสดงปุ่มที่คลิก, แสดงคีย์ที่กด บนหน้าต่างหรือ log)
โดยไม่ได้ยิงคำสั่งเข้าระบบปฏิบัติการจริง

**ใช้สำหรับ:** สาธิตหลักการทำงานเบื้องหลังของโปรแกรมควบคุมระยะไกล (remote desktop /
screen sharing) เฉพาะส่วนช่องทางอินพุต, เป็นฐานสำหรับ remote presenter (ชี้ pointer
บนสไลด์ของเครื่องอื่น) หรือขยายเป็นห้องที่หลายคนเห็น cursor ของกันและกันแบบเรียลไทม์

**ขอบเขต (สำคัญ):** โปรแกรมนี้เน้นที่ *การออกแบบโพรโทคอลการส่งอินพุต* จึงตัดส่วน
การจับภาพหน้าจอ (framebuffer streaming) และการยิงอินพุตเข้า OS จริงออกไป — ฝั่ง Viewer
แสดงผลแบบจำลองเท่านั้น ทำให้โฟกัสอยู่ที่กลไกเครือข่ายซึ่งเป็นหัวใจของวิชานี้

---

## 2. คุณลักษณะของแอปพลิเคชัน (Application Characteristics)

| ด้าน | ความต้องการของ RICP | เหตุผล |
|---|---|---|
| **Reliable data transfer** | **ทนต่อการสูญหายได้ (loss-tolerant) สำหรับ MOVE** แต่ **ต้องเชื่อถือได้สำหรับ CLICK/KEY** | ตำแหน่งเมาส์ที่หายไปไม่ต้องส่งซ้ำ เพราะตำแหน่งใหม่ที่สดกว่ากำลังจะมาถึง แต่การคลิกหรือกดคีย์ที่หายไปทำให้ความหมายเปลี่ยน |
| **Throughput** | **ต่ำถึงปานกลาง (elastic)** ~1–10 KB/s | แต่ละ event มีขนาดเล็ก (6–12 ไบต์) ความถี่สูง (สูงสุด ~120 event/วินาที) |
| **Timing / Delay** | **ต้องการ delay ต่ำมาก (< 50–100 ms)** | เป็นหัวใจของแอปพลิเคชัน — ถ้า pointer ตามช้าจะรู้สึกหน่วงและใช้ควบคุมไม่ได้ |
| **Security** | **พื้นฐาน** — มี session token ตอน handshake | ป้องกันการส่ง event มั่วจากที่อื่น (ยังไม่เข้ารหัสช่องทาง) |

**คุณลักษณะเชิงโครงสร้าง**

1. **Latency-critical** — คุณค่าของ event ลดลงตามเวลา event ที่มาช้าอาจไร้ประโยชน์
2. **High-frequency, small-payload** — event ถี่และเล็ก overhead ต่อ packet จึงสำคัญมาก
3. **Mixed reliability** — ข้อมูลต่างชนิดต้องการความเชื่อถือต่างกัน (MOVE ทิ้งได้, CLICK/KEY ทิ้งไม่ได้)
4. **Unidirectional data, bidirectional control** — ข้อมูลอินพุตไหลทางเดียว (Controller→Viewer)
   แต่มีข้อความควบคุม (handshake, ack, stats) ไหลกลับ
5. **Binary, fixed-length messages** — ต่างจากโพรโทคอลแบบข้อความ เพื่อลด overhead และ parse เร็ว

---

## 3. เหตุผลในการเลือก Transport Layer Service Model: **UDP**

RICP เลือก **UDP** เป็นช่องทางหลัก ด้วยเหตุผล 4 ข้อ (และใช้ TCP เสริมเฉพาะบางส่วน — ดูหัวข้อ 3.5)

**3.1 ต้องการ Delay ต่ำที่สุด — หลีกเลี่ยง Head-of-Line Blocking**
นี่คือเหตุผลหลัก TCP รับประกันการส่งแบบครบและเรียงลำดับ ด้วยการ *กัก* ข้อมูลที่มาถึงแล้ว
ไว้จนกว่าข้อมูลก่อนหน้าที่หายไปจะถูกส่งซ้ำสำเร็จ (head-of-line blocking) สำหรับ pointer
สิ่งนี้เป็นหายนะ — ถ้า MOVE อันที่ 42 หาย TCP จะกัก MOVE อันที่ 43, 44, 45 ไว้ทั้งหมด
เพื่อรอส่ง 42 ซ้ำ ทำให้ pointer **ค้าง** แล้วกระโดด แทนที่จะเลื่อนลื่น UDP ส่งตรงไม่กัก
pointer จึงลื่นแม้มี loss

**3.2 ข้อมูลที่หายไปไม่มีค่าพอที่จะส่งซ้ำ**
กลไก retransmission ของ TCP ออกแบบมาเพื่อข้อมูลที่ *ต้อง* ครบ แต่ตำแหน่งเมาส์เก่า
เมื่อส่งซ้ำสำเร็จ ตำแหน่งจริงก็เปลี่ยนไปหลายรอบแล้ว การส่งซ้ำจึงเปลืองเปล่าและยัง
เพิ่ม latency ให้ข้อมูลใหม่ UDP ที่ไม่ส่งซ้ำจึงเหมาะกว่าโดยธรรมชาติของข้อมูล

**3.3 Overhead ต่อ packet ต่ำกว่า**
event มีขนาด 6–12 ไบต์ แต่ TCP header มี 20 ไบต์ + ต้องมีการ ACK ทุก segment ขณะที่
UDP header เพียง 8 ไบต์และไม่มี ACK ในตัว เมื่อส่ง event ถี่ ๆ ความประหยัดนี้มีนัยสำคัญ

**3.4 ไม่ต้องการ connection setup**
TCP ต้อง 3-way handshake ก่อนเริ่มส่ง UDP ส่งได้ทันที เหมาะกับ event ที่ต้องเริ่มไว
(RICP มี handshake ระดับแอปพลิเคชันของตัวเองที่เบากว่า — ดูหัวข้อ 6)

**3.5 สิ่งที่ต้องชดเชยเอง (trade-off ที่ยอมรับ)**
UDP ไม่รับประกัน (ก) การส่งถึง (ข) ลำดับ (ค) การไม่ซ้ำ RICP จึงเพิ่มกลไกที่ชั้น
แอปพลิเคชันเอง: **sequence number** (แก้ ข, ค และตรวจ ก), **coalescing** (ลดผลกระทบจาก
การส่งไม่ทัน), และสำหรับ **CLICK/KEY ที่ทิ้งไม่ได้** ใช้กลไก **application-level ACK +
retransmit** หรือส่งผ่าน **ช่อง TCP คู่ขนาน** (ทางเลือกในหัวข้อ 9) — เท่ากับเลือกเติม
เฉพาะการรับประกันที่จำเป็น แทนที่จะรับ overhead ของ TCP มาทั้งก้อน

**สรุปเปรียบเทียบ**

| | ถ้าใช้ TCP | ถ้าใช้ UDP (ที่เลือก) |
|---|---|---|
| MOVE หาย 1 packet | pointer ค้างรอ retransmit | ข้ามไปตำแหน่งใหม่ ลื่นต่อ |
| Latency | สูงขึ้นเมื่อมี loss | คงที่ ไม่ขึ้นกับ loss |
| Overhead/packet | 20B header + ACK | 8B header |
| ความครบถ้วน | ครบเสมอ | ต้องจัดการเอง |

**กรณีที่ควรเลือก TCP แทน:** ถ้าเป็นการส่ง "ชุดคำสั่งมาโคร" ที่ทุก event ต้องถึงครบและ
เรียงเป๊ะ (เล่นซ้ำอัตโนมัติ) ความครบถ้วนสำคัญกว่า latency → TCP เหมาะกว่า

---

## 4. ภาพรวมโพรโทคอล

| หัวข้อ | ค่า |
|---|---|
| ชื่อ | RICP — Remote Input Control Protocol |
| เวอร์ชัน | RICP/1.0 |
| Transport หลัก | UDP พอร์ต 9200 (ข้อมูล input) |
| Transport เสริม (ทางเลือก) | TCP พอร์ต 9201 (handshake ที่เชื่อถือได้ + CLICK/KEY) |
| Endianness | Big-endian (network byte order) |
| รูปแบบข้อความ | Binary, fixed-length ต่อชนิด |
| ขนาด datagram สูงสุด | 512 ไบต์ (ต่ำกว่า MTU ปลอดภัย ไม่เกิด IP fragmentation) |

### 4.1 หลักการ Framing บน UDP

ต่างจาก TCP ที่เป็น byte stream (ต้องมี delimiter หรือ length-prefix เพื่อรู้ขอบเขต
ข้อความ) — **UDP ส่งเป็น datagram: 1 การส่ง = 1 การรับ เสมอ** ขอบเขตข้อความจึงถูก
รักษาไว้โดยตัว transport เอง ผู้รับอ่านหนึ่ง datagram ได้หนึ่ง message ครบพอดี
ไม่มีปัญหาข้อความติดกันหรือขาดครึ่งแบบ TCP

RICP ใช้ประโยชน์จากข้อนี้ด้วยการทำ message เป็น **fixed-length ต่อชนิด** — ผู้รับดู
byte แรก (type) ก็รู้ทันทีว่า datagram นี้ยาวกี่ไบต์และมี field อะไรบ้าง

---

## 5. รูปแบบข้อความ (Message Format)

ทุก datagram ขึ้นต้นด้วย **header 5 ไบต์** ตามด้วย payload ที่ขึ้นกับชนิด

### 5.1 Common Header (5 ไบต์)

```
+--------+--------+--------+--------+--------+
|         seq (U32, big-endian)     |  type  |
|            4 bytes                 | 1 byte |
+--------+--------+--------+--------+--------+
```

- `seq` — เลขลำดับ เพิ่มทีละ 1 ทุก datagram ที่ Controller ส่ง (เริ่มที่ค่าที่ตกลงตอน handshake)
- `type` — ชนิดข้อความ (ตารางด้านล่าง)

### 5.2 ชนิดข้อความและ payload

| type | ชื่อ | ทิศทาง | ขนาดรวม | payload |
|---|---|---|---|---|
| 0x01 | MOVE | C→V | 10 B | x (U16), y (U16), button-mask (U8) |
| 0x02 | CLICK | C→V | 8 B | button-mask (U8), down-flag (U8), reserved (U8) |
| 0x03 | KEY | C→V | 11 B | down-flag (U8), reserved (U8), keysym (U32) |
| 0x04 | SCROLL | C→V | 8 B | delta-x (S8), delta-y (S8), reserved (U8) |
| 0x10 | HELLO | C→V | 5 B + token | เจรจาเริ่ม session |
| 0x11 | WELCOME | V→C | 9 B | start-seq (U32) |
| 0x12 | ACK | V→C | 9 B | ack-seq (U32) — ยืนยันรับ CLICK/KEY |
| 0x13 | STATS | V→C | 17 B | received (U32), lost (U32), reordered (U32) |
| 0x20 | STATUS | V→C | 8 B + phrase | code (U16), phraseLen (U8), phrase (ข้อความ) |
| 0x1F | BYE | C→V | 5 B | ปิด session |

**MOVE (0x01) — 10 ไบต์** — event ที่ถี่ที่สุด:
```
[ seq(4) ][ type=0x01 ][ x(2) ][ y(2) ][ button-mask(1) ]
```
`button-mask` เก็บสถานะปุ่มปัจจุบันแบบ bitmask (ยืมจาก RFB): bit0=ซ้าย, bit1=กลาง,
bit2=ขวา (0=ปล่อย 1=กด) ทำให้ MOVE พก "ตำแหน่ง + สถานะปุ่มขณะลาก" ไปพร้อมกันได้

**KEY (0x03) — 11 ไบต์**:
```
[ seq(4) ][ type=0x03 ][ down-flag(1) ][ reserved(1) ][ keysym(4) ]
```
`down-flag`: 1=กด 0=ปล่อย (แยก event กด/ปล่อย ตามแบบ RFB → รองรับกดค้าง, คีย์ผสม)
`keysym`: ใช้ค่า keysym ของ X11 (RFC 6143) — ตัวอักษรทั่วไป = ASCII (`A`=0x41),
ปุ่มพิเศษ: Enter=0xFF0D, Esc=0xFF1B, Backspace=0xFF08, ลูกศรซ้าย=0xFF51, F1=0xFFBE,
Shift ซ้าย=0xFFE1, Ctrl ซ้าย=0xFFE3

**STATUS (0x20) — ความยาวแปรผัน** — ข้อความควบคุมที่พก status code + คำอธิบาย:
```
[ seq(4) ][ type=0x20 ][ code(2) ][ phraseLen(1) ][ phrase(ยาวตาม phraseLen) ]
```
`code` เป็น U16 (2 ไบต์) เพราะรหัสวิ่ง 100–599 (เกิน 255 แต่ไม่ถึงพันล้าน)
`phraseLen` บอกความยาวของ phrase เพื่อให้ผู้รับรู้ว่าต้องอ่านกี่ไบต์ (แบบเดียวกับ HELLO)
ตัว phrase ส่งมาเพื่อให้ log อ่านออกโดยไม่ต้องมีตาราง code ทั้งสองฝั่ง

### 5.3 รหัสสถานะ (Status Codes)

STATUS ใช้รหัสตัวเลข 3 หลักแบ่งกลุ่มตามหลักเดียวกับ HTTP เพื่อให้ผู้รับตัดสินใจได้
จากหลักแรกแม้ไม่รู้จักรหัสนั้น รหัสถูกส่งผ่านข้อความ STATUS (0x20) ที่ Viewer ตอบกลับ

**1xx — Informational**

| code | phrase | ส่งเมื่อ |
|---|---|---|
| 100 | STREAMING | Viewer เริ่มรับ event stream แล้ว |

**2xx — Success**

| code | phrase | ส่งเมื่อ |
|---|---|---|
| 200 | WELCOME | handshake สำเร็จ (ใช้คู่กับ/แทน WELCOME message) |
| 202 | ACK | ยืนยันรับ CLICK/KEY ที่ต้องเชื่อถือได้ |
| 210 | BYE OK | ปิด session ตามคำสั่ง |

**4xx — Client error (Controller ผิด)**

| code | phrase | ส่งเมื่อ |
|---|---|---|
| 400 | BAD DATAGRAM | ขนาด/รูปแบบ datagram ผิด (สั้นกว่า header, type ไม่รู้จัก) |
| 401 | UNAUTHORIZED | ส่ง event มาก่อน handshake / token ผิด |
| 408 | HANDSHAKE TIMEOUT | HELLO มาช้าเกินกำหนด |
| 413 | DATAGRAM TOO LARGE | เกิน 512 ไบต์ |
| 426 | VERSION NOT SUPPORTED | เวอร์ชันไม่ตรง (RICP/1.0) |

**5xx — Server/Viewer error**

| code | phrase | ส่งเมื่อ |
|---|---|---|
| 500 | INTERNAL ERROR | ข้อผิดพลาดภายในฝั่ง Viewer |
| 503 | VIEWER BUSY | Viewer รับไม่ไหว/เต็ม |

> หมายเหตุ: packet loss และ reordering **ไม่ใช่** สถานะผิดพลาด (ไม่ใช่ 4xx/5xx)
> แต่เป็นเรื่องปกติของ UDP ที่คาดไว้แล้ว จึงรายงานเป็น *ตัวเลขสถิติ* ผ่าน STATS (0x13)
> ไม่ใช่ผ่าน STATUS

---

## 6. กลไกหลักของโพรโทคอล

### 6.1 Sequence number — ตรวจจับ loss และ reorder
ทุก datagram พก `seq` ที่เพิ่มทีละ 1 ฝั่ง Viewer เก็บ `lastSeq` ที่เห็นล่าสุด แล้วเทียบ:

- ได้ seq = lastSeq + 1 → ปกติ
- ได้ seq > lastSeq + 1 → **หาย** (seq - lastSeq - 1) packet → เพิ่มตัวนับ lost
- ได้ seq ≤ lastSeq → **มาช้า/ซ้ำ** → นับ reordered แล้ว **ทิ้ง** (สำหรับ MOVE)

### 6.2 Coalescing — รวม MOVE ที่ค้างคิว
ถ้าอัตราการเกิด event สูงกว่าอัตราที่ส่งทัน Controller จะ **รวม MOVE หลายอันเป็นอันล่าสุด**
ก่อนส่ง (ส่งเฉพาะตำแหน่งปัจจุบัน ไม่ส่งทุกก้าวระหว่างทาง) — ใช้ได้เฉพาะ MOVE เท่านั้น
CLICK/KEY/SCROLL ห้ามรวมหรือทิ้ง เพราะทุกครั้งมีความหมายเฉพาะ

### 6.3 Drop-if-older — เลือกความสดใหม่
ฝั่ง Viewer วาดเฉพาะ MOVE ที่ seq ใหม่กว่าที่วาดล่าสุด ถ้า datagram มาช้ากว่าที่วาดไปแล้ว
จะทิ้ง เพราะการวาดตำแหน่งเก่าทับตำแหน่งใหม่ทำให้ pointer กระตุกย้อนหลัง

### 6.4 Reliable path สำหรับ CLICK/KEY (ทางเลือก — stretch)
CLICK/KEY ทิ้งไม่ได้ RICP รองรับ 2 แนวทาง เลือกอย่างใดอย่างหนึ่ง:
- **(ก) App-level ACK บน UDP:** Viewer ตอบ `ACK ack-seq` เมื่อรับ CLICK/KEY; Controller
  เก็บ event ที่ยังไม่ถูก ack ไว้ ถ้าไม่ได้ ack ภายใน timeout (เช่น 100 ms) ส่งซ้ำ
- **(ข) ช่อง TCP คู่ขนาน:** ส่ง CLICK/KEY ผ่าน TCP (พอร์ต 9201) ส่วน MOVE ผ่าน UDP —
  ได้ความเชื่อถือของ TCP เฉพาะข้อมูลที่ต้องการ โดยไม่กระทบ latency ของ MOVE
  (แนวทางเดียวกับ game streaming ที่แยก traffic ตามความสำคัญ)

### 6.5 STATS — วัดผล
Viewer ส่ง STATS กลับเป็นระยะ (เช่นทุก 1 วินาที) รายงาน received / lost / reordered
เพื่อให้ demo แสดงคุณภาพการส่งเป็นตัวเลขได้จริง

---

## 7. เครื่องสถานะของเซสชัน (Session State Machine)

```
    Controller                         Viewer
        |                                |
        |------ HELLO(token) ----------->|   ตรวจ token
        |<----- WELCOME(start-seq) ------|   ตอบเลขเริ่ม
        |                                |
    [ ACTIVE ]                       [ ACTIVE ]
        |                                |
        |== MOVE/CLICK/KEY/SCROLL =====>>|   (stream, seq เพิ่มเรื่อย ๆ)
        |<----- ACK / STATS -------------|   (control กลับ)
        |                                |
        |------ BYE -------------------->|   ปิด session
        v                                v
    [ CLOSED ]                       [ CLOSED ]
```

- **INIT → ACTIVE:** ต้อง HELLO/WELCOME สำเร็จก่อน จึงจะรับ event (event ที่มาก่อน
  handshake จะถูกทิ้ง)
- **timeout:** ถ้า Viewer ไม่ได้รับ datagram ใด ๆ เกิน N วินาที ถือว่า session ตาย
  (เพราะ UDP ไม่มีสัญญาณตัดการเชื่อมต่อแบบ TCP)

---

## 8. ตัวอย่างการสื่อสาร

`C:` = Controller, `V:` = Viewer (แสดงเป็นเชิงสัญลักษณ์)

```
C: HELLO token=abc123
V: WELCOME start-seq=1000

C: seq=1000 MOVE  (100,100) mask=0
C: seq=1001 MOVE  (104,103) mask=0
C: seq=1002 MOVE  (110,108) mask=0
       [seq 1003 หายระหว่างทาง]
C: seq=1004 MOVE  (121,119) mask=0
V:   -> ตรวจพบ: หาย 1 (คาด 1003 ได้ 1004) นับ lost+1, วาด 1004 ต่อได้เลย

C: seq=1005 CLICK left down
V:   -> วาดคลิก, ตอบ ACK ack-seq=1005
C: seq=1006 CLICK left up
V:   -> ACK ack-seq=1006

C: seq=1007 KEY  down 0x41 ('A')
C: seq=1008 KEY  up   0x41
V:   -> แสดง "A", ACK ทั้งสอง

       [seq 1002 เดินทางช้า เพิ่งมาถึงหลัง 1004]
V:   -> seq 1002 ≤ lastSeq(1008): นับ reordered+1, ทิ้ง

V: STATS received=8 lost=1 reordered=1

C: BYE
```

---

## 9. Related Work

RICP ได้แรงบันดาลใจจากสองแหล่ง

**RFB/VNC (RFC 6143)** — ยืมโครง input message: byte แรกเป็น message-type,
message แบบ fixed-length, การเก็บสถานะปุ่มเมาส์เป็น `button-mask` แบบ bitmask,
การแยก key press/release ด้วย `down-flag`, และการใช้ค่า keysym ของ X11 สำหรับปุ่มพิเศษ
ความต่างสำคัญคือ RFB วิ่งบน TCP จึง**ไม่มี sequence number ในข้อความ** (พึ่ง in-order
delivery ของ TCP) ขณะที่ RICP วิ่งบน UDP จึงต้องเพิ่ม seq เองเพื่อตรวจจับ loss/reorder
และ RICP ตัด framebuffer streaming ของ RFB ออก เหลือเฉพาะช่องอินพุต

**Game/Cloud streaming netcode (เช่น Parsec BUD)** — ยืมหลักการเลือก UDP เพื่อความลื่นไหล
แทนความครบถ้วน, การใส่ sequence/timestamp ในทุก input packet, และแนวคิดแยก traffic
ตามความสำคัญ (real-time บน UDP, ข้อมูลที่ห้ามหายบนช่องที่เชื่อถือได้) ซึ่งสะท้อนใน
reliable path ของ RICP (หัวข้อ 6.4)

---

## 10. ข้อจำกัดและแนวทางพัฒนาต่อ

- ยังไม่เข้ารหัสช่องทาง (production ควรใช้ DTLS เหมือน Parsec)
- token ตอน handshake เป็นการยืนยันตัวตนอย่างง่าย ยังไม่กันการดักจับ/เล่นซ้ำ (replay)
- แสดงผลแบบจำลอง ไม่ยิงอินพุตเข้า OS จริง (โดยตั้งใจ เพื่อโฟกัสที่โพรโทคอล)
- Stretch: โหมด broadcast — Viewer หลายเครื่อง / หลายคนเห็น cursor ของกันและกัน
  ผ่าน server ตรงกลาง (ต่อยอดเป็น "cursor arena")
