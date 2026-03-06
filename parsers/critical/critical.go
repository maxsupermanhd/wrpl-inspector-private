package critical

import (
	"encoding/binary"
	"main/idfieldserializer"

	"github.com/maxsupermanhd/wrpl-inspector/v2/wrpl/danet"
	"github.com/maxsupermanhd/wrpl-inspector/v2/wrpl/packet"
)

type CriticalDamageParser struct {
}

func (p *CriticalDamageParser) Name() string {
	return "critical"
}

func (p *CriticalDamageParser) ParsesMatching() map[byte][][]packet.ParsingCondition {
	return map[byte][][]packet.ParsingCondition{
		4: {{
			packet.NewParsingCondition(0, 0x02),
			packet.NewParsingCondition(1, 0x58),
			packet.NewParsingCondition(2, 0x56),
			packet.NewParsingCondition(3, 0xf0),
		}},
	}
}

type CriticalDamagePacket struct {
	OffendedUID uint16
	PlayerPID   uint32
	Vehicle     string
	PlayerUID   uint16
	Fire        bool
	Unk0        byte
}

func (p *CriticalDamageParser) Parse(pk *packet.Packet) (any, error) {
	ret := CriticalDamagePacket{}
	r := danet.NewBitReader(pk.PacketPayload[4:])
	err := idfieldserializer.DeserializeIdFieldSerializer32(r, func(fieldNum uint8) error {
		var err error
		switch fieldNum {
		case 1:
			return binary.Read(r, binary.LittleEndian, &ret.OffendedUID)
		case 2:
			return binary.Read(r, binary.LittleEndian, &ret.PlayerPID)
		case 3:
			return r.ReadLenStrInto(&ret.Vehicle)
		case 4:
			return binary.Read(r, binary.LittleEndian, &ret.PlayerUID)
		case 5:
			return r.ReadBoolInto(&ret.Fire)
		case 6:
			ret.Unk0, err = r.ReadByte()
		default:
			return idfieldserializer.ErrSkipField
		}
		return err
	})
	return ret, err
}
