package paths

import (
	"encoding/binary"
	"math"
	"wrplinspectorprivate/game"

	"github.com/maxsupermanhd/wrpl-inspector/v2/wrpl/danet"
	"github.com/maxsupermanhd/wrpl-inspector/v2/wrpl/packet"
)

type PositionRetainerParser struct {
	Paths map[uint64][]game.SpaceTime
}

func NewPositionRetainerParser() *PositionRetainerParser {
	return &PositionRetainerParser{
		Paths: map[uint64][]game.SpaceTime{},
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
	st := game.SpaceTime{
		Time: pk.CurrentTime,
		X:    math.Float64frombits(binary.LittleEndian.Uint64(pk.PacketPayload[14:])),
		Y:    math.Float64frombits(binary.LittleEndian.Uint64(pk.PacketPayload[22:])),
		Z:    math.Float64frombits(binary.LittleEndian.Uint64(pk.PacketPayload[30:])),
	}
	p.Paths[eid] = append(p.Paths[eid], st)
	return st, nil
}
