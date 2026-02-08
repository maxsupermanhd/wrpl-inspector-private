package stub1

import (
	"github.com/maxsupermanhd/wrpl-inspector/v2/wrpl/danet"
	"github.com/maxsupermanhd/wrpl-inspector/v2/wrpl/packet"
)

type StubData struct {
	EID uint64
	F   [10]float32
}

type PacketStubParser struct{}

func (p *PacketStubParser) Name() string {
	return "stub1"
}

func (p *PacketStubParser) ParsesMatching() map[byte][][]packet.ParsingCondition {
	return map[byte][][]packet.ParsingCondition{
		4: {{
			packet.NewParsingCondition(0, 0x02),
			packet.NewParsingCondition(1, 0x58),
			packet.NewParsingCondition(2, 0x74),
			packet.NewParsingCondition(3, 0xf0),
		}},
	}
}

func (p *PacketStubParser) Parse(pk *packet.Packet) (any, error) {
	if len(pk.PacketPayload) < 40 {
		return nil, nil
	}
	parsed := &StubData{}
	var err error
	parsed.EID, err = danet.NewBitReader(pk.PacketPayload[4:]).ReadCompressed()
	if err != nil {
		return nil, err
	}
	return parsed, nil
}
