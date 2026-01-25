package kills2

import (
	"encoding/binary"
	"fmt"

	"github.com/maxsupermanhd/wrpl-inspector/wrpl/danet"
	"github.com/maxsupermanhd/wrpl-inspector/wrpl/packet"
)

type KillEntry struct {
	Seq         uint64
	CurrentTime uint32

	Control       byte
	KillerID      byte
	KillerVehicle string
	KillerUID     uint16
	VictimUID     uint16
	Weapon        string
	VictimID      uint32
}

type PacketKillParser struct {
	KeepKills bool
	Kills     []KillEntry
}

func (p *PacketKillParser) Name() string {
	return "kill2"
}

func (p *PacketKillParser) ParsesMatching() map[byte][][]packet.ParsingCondition {
	return map[byte][][]packet.ParsingCondition{
		4: {{
			packet.NewParsingCondition(0, 0x02),
			packet.NewParsingCondition(1, 0x58),
			packet.NewParsingCondition(2, 0x58),
			packet.NewParsingCondition(3, 0xf0),
		}},
	}
}

func (p *PacketKillParser) Parse(pk *packet.Packet) (any, error) {
	parsed := KillEntry{
		Seq:         pk.Seq,
		CurrentTime: pk.CurrentTime,
	}
	var err error
	r := danet.NewBitReader(pk.PacketPayload)
	parsed.Control, err = r.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("reading control: %w", err)
	}
	r.IgnoreBytes(3) // always 0x00FE3F
	parsed.KillerID, err = r.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("reading killer id: %w", err)
	}
	r.IgnoreBytes(3) // 0x000000
	parsed.KillerVehicle, err = r.ReadLenStr()
	if err != nil {
		return nil, fmt.Errorf("reading killer vehicle: %w", err)
	}
	err = binary.Read(r, binary.LittleEndian, &parsed.KillerUID)
	if err != nil {
		return nil, fmt.Errorf("reading killer uid: %w", err)
	}
	err = binary.Read(r, binary.LittleEndian, &parsed.VictimUID)
	if err != nil {
		return nil, fmt.Errorf("reading victim uid: %w", err)
	}
	r.IgnoreBits(1) // unk bool
	r.IgnoreBits(8) // unk uint8
	r.IgnoreBits(8) // unk uint8
	parsed.Weapon, err = r.ReadLenStr()
	if err != nil {
		return nil, fmt.Errorf("reading weapon: %w", err)
	}
	err = binary.Read(r, binary.LittleEndian, &parsed.VictimID)
	if err != nil {
		return nil, fmt.Errorf("reading victim id: %w", err)
	}
	if p.KeepKills {
		p.Kills = append(p.Kills, parsed)
	}
	return parsed, nil
}
