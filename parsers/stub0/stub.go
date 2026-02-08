package stub0

import (
	"encoding/binary"
	"math"

	"github.com/maxsupermanhd/wrpl-inspector/v2/wrpl/danet"
	"github.com/maxsupermanhd/wrpl-inspector/v2/wrpl/packet"
)

type StubData struct {
	CurrentTime uint32
	EID         uint64
	F           [10]float32
	Rest        [5]byte
}

type PacketStubParser struct {
	Data map[uint64][]StubData
}

func (p *PacketStubParser) Name() string {
	return "stub0"
}

func (p *PacketStubParser) ParsesMatching() map[byte][][]packet.ParsingCondition {
	return map[byte][][]packet.ParsingCondition{
		4: {{
			packet.NewParsingCondition(0, 0xff),
			packet.NewParsingCondition(1, 0x0f),
			packet.NewParsingCondition(5, 0xcc),
			packet.NewParsingCondition(6, 0xf0),
			packet.NewParsingCondition(7, 0x34),
			packet.NewParsingCondition(8, 0x00),
			packet.NewParsingCondition(9, 0xfe),
			packet.NewParsingCondition(10, 0x01),
		}, {
			packet.NewParsingCondition(0, 0xff),
			packet.NewParsingCondition(1, 0x07),
			packet.NewParsingCondition(5, 0xcc),
			packet.NewParsingCondition(6, 0xf0),
			packet.NewParsingCondition(7, 0x34),
			packet.NewParsingCondition(8, 0x00),
			packet.NewParsingCondition(9, 0xfe),
			packet.NewParsingCondition(10, 0x01),
		}},
		//  0 1 2 3 4 5 6 7 8 9 a
		// ^ff0f......ccf03400fe01
	}
}

func (p *PacketStubParser) Parse(pk *packet.Packet) (any, error) {
	if len(pk.PacketPayload) < 40 {
		return nil, nil
	}
	var err error
	eid, err := danet.NewBitReader(pk.PacketPayload[2:]).ReadCompressed()
	if err != nil {
		return nil, err
	}
	parsed := StubData{
		CurrentTime: pk.CurrentTime,
		F:           [10]float32{},
	}
	parsed.F[0] = math.Float32frombits(binary.LittleEndian.Uint32(pk.PacketPayload[11+4*0:]))
	parsed.F[1] = math.Float32frombits(binary.LittleEndian.Uint32(pk.PacketPayload[11+4*1:]))
	parsed.F[2] = math.Float32frombits(binary.LittleEndian.Uint32(pk.PacketPayload[11+4*2:]))
	parsed.F[3] = math.Float32frombits(binary.LittleEndian.Uint32(pk.PacketPayload[11+4*3:]))
	parsed.F[4] = math.Float32frombits(binary.LittleEndian.Uint32(pk.PacketPayload[11+4*4:]))
	parsed.F[5] = math.Float32frombits(binary.LittleEndian.Uint32(pk.PacketPayload[11+4*5:]))
	parsed.F[6] = math.Float32frombits(binary.LittleEndian.Uint32(pk.PacketPayload[11+4*6:]))
	parsed.F[7] = math.Float32frombits(binary.LittleEndian.Uint32(pk.PacketPayload[11+4*7:]))
	parsed.F[8] = math.Float32frombits(binary.LittleEndian.Uint32(pk.PacketPayload[11+4*8:]))
	parsed.F[9] = math.Float32frombits(binary.LittleEndian.Uint32(pk.PacketPayload[11+4*9:]))
	// 18 - normal
	// 21 - scope
	// 1d - binocular
	parsed.Rest[0] = pk.PacketPayload[11+4*10+0]
	parsed.Rest[1] = pk.PacketPayload[11+4*10+1]
	parsed.Rest[2] = pk.PacketPayload[11+4*10+2]
	parsed.Rest[3] = pk.PacketPayload[11+4*10+3]
	parsed.Rest[4] = pk.PacketPayload[11+4*10+4]
	parsed.EID = eid
	p.Data[eid] = append(p.Data[eid], parsed)
	return parsed, nil
}
