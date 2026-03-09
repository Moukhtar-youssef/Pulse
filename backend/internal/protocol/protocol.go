package protocol

const (
	Magic   uint16 = 0x5055
	Version uint8  = 1

	TypeHello     uint8 = 1
	TypeHeartbeat uint8 = 2
	TypeMetrics   uint8 = 3
	TypeAlert     uint8 = 4
	TypeLog       uint8 = 5
)

type Header struct {
	Magic   uint16
	Version uint8
	Type    uint8
	Length  uint32
}

type Packet struct {
	Type    uint8
	Payload []byte
}
