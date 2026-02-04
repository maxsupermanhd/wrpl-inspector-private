package paths

import (
	"encoding/binary"
	"math"

	"github.com/maxsupermanhd/wrpl-inspector/v2/wrpl/danet"
	"github.com/maxsupermanhd/wrpl-inspector/v2/wrpl/packet"
)

type SpaceTime struct {
	Time uint32
	X    int64
	Y    int64
	Z    int64
}

type PositionRetainerParser struct {
	Paths map[uint64][]SpaceTime
}

func NewPositionRetainerParser() *PositionRetainerParser {
	return &PositionRetainerParser{
		Paths: map[uint64][]SpaceTime{},
	}
}

func (p *PositionRetainerParser) Name() string {
	return "paths"
}

func (p *PositionRetainerParser) ParsesMatching() map[byte][][]packet.ParsingCondition {
	return map[byte][][]packet.ParsingCondition{
		4: {{
			packet.NewParsingCondition(0, 0xff),
			packet.NewParsingCondition(1, 0x0f),
			packet.NewParsingCondition(5, 0xa3),
			packet.NewParsingCondition(6, 0xf0),
			packet.NewParsingCondition(10, 0x00),
			packet.NewParsingCondition(11, 0x00),
			packet.NewParsingCondition(13, 0x14),
		}},
	}
}

func (p *PositionRetainerParser) Parse(pk *packet.Packet) (any, error) {
	if len(pk.PacketPayload) < 40 {
		return nil, nil
	}
	eid, err := danet.NewBitReader(pk.PacketPayload[2:]).ReadCompressed()
	if err != nil {
		return nil, err
	}
	st := SpaceTime{
		Time: pk.CurrentTime,
		X:    int64(math.Float64frombits(binary.LittleEndian.Uint64(pk.PacketPayload[14:]))),
		Y:    int64(math.Float64frombits(binary.LittleEndian.Uint64(pk.PacketPayload[22:]))),
		Z:    int64(math.Float64frombits(binary.LittleEndian.Uint64(pk.PacketPayload[30:]))),
	}
	p.Paths[eid] = append(p.Paths[eid], st)
	return nil, nil
}

func (p *PositionRetainerParser) LastPosition(eid uint64) *SpaceTime {
	s := p.Paths[eid]
	if len(s) == 0 {
		return nil
	}
	return &s[len(s)-1]
}
