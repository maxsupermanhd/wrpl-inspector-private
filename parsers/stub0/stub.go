package stub0

import (
	"bytes"
	"encoding/binary"

	"github.com/maxsupermanhd/wrpl-inspector/wrpl/danet"
	"github.com/maxsupermanhd/wrpl-inspector/wrpl/packet"
)

type StubData struct {
	EID uint64
	F   [10]float32
}

type PacketStubParser struct{}

func (p *PacketStubParser) Name() string {
	return "stub"
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
	parsed := &StubData{}
	var err error
	parsed.EID, err = danet.NewBitReader(pk.PacketPayload[2:]).ReadCompressed()
	if err != nil {
		return nil, err
	}
	r := bytes.NewReader(pk.PacketPayload[11:])
	binary.Read(r, binary.LittleEndian, &parsed.F[0])
	binary.Read(r, binary.LittleEndian, &parsed.F[1])
	binary.Read(r, binary.LittleEndian, &parsed.F[2])
	binary.Read(r, binary.LittleEndian, &parsed.F[3])
	binary.Read(r, binary.LittleEndian, &parsed.F[4])
	binary.Read(r, binary.LittleEndian, &parsed.F[5])
	binary.Read(r, binary.LittleEndian, &parsed.F[6])
	binary.Read(r, binary.LittleEndian, &parsed.F[7])
	binary.Read(r, binary.LittleEndian, &parsed.F[8])
	binary.Read(r, binary.LittleEndian, &parsed.F[9])
	return parsed, nil
}
