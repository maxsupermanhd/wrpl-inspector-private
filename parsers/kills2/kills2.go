package kills2

import (
	"encoding/binary"
	"fmt"
	"main/parsers/ecs2"
	"main/parsers/paths"

	"github.com/maxsupermanhd/wrpl-inspector/v2/wrpl/danet"
	"github.com/maxsupermanhd/wrpl-inspector/v2/wrpl/packet"
)

type KillEntry struct {
	Seq         uint64
	CurrentTime uint32

	KillerPid              uint32
	KillerUid              uint16
	ResolvedKiller         *ecs2.Entity
	ResolvedKillerPosition *paths.SpaceTime

	VictimPid              uint32
	VictimUid              uint16
	ResolvedVictim         *ecs2.Entity
	ResolvedVictimPosition *paths.SpaceTime

	PlayerVehicle   string
	PlayerWeapon    string
	DestroyedWeapon string
}

type PacketKillParser struct {
	KeepKills bool
	Kills     []KillEntry
	ECS       *ecs2.EntityManager
	Paths     *paths.PositionRetainerParser
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
	var serializer = ecs2.IdFieldSerializer32{}
	r.IgnoreBytes(4)
	fields, err := serializer.ReadFieldsSizeAndFlag(r)
	if err != nil {
		return nil, err
	}
	var index uint8
	for fields > 0 {
		var uVar3 uint8
		for fields>>uVar3&1 == 0 {
			uVar3++
		}
		fields = fields & ^(1 << (uVar3 & 0x1f))
		switch uVar3 {
		case 1:
			err = binary.Read(r, binary.LittleEndian, &parsed.KillerPid)
			if err != nil {
				return nil, err
			}
		case 2:
			parsed.PlayerVehicle, err = r.ReadLenStr()
			if err != nil {
				return nil, err
			}
		case 3:
			err = binary.Read(r, binary.LittleEndian, &parsed.VictimUid)
			parsed.VictimUid &= 0x7FF
			if err != nil {
				return nil, err
			}
			if parsed.VictimUid != 0x7FF {
				resolved, ok := p.ECS.GetEntityByUid(int32(parsed.VictimUid))
				if !ok {
					return nil, fmt.Errorf("failed to resolve victim uid %v", parsed.VictimUid)
				}
				parsed.ResolvedVictim = resolved
				if p.Paths != nil {
					for eid, pv := range p.Paths.Paths {
						eid2 := ((uint64(uint64(eid)&0xff) << uint64(0x16)) | (uint64(eid) >> uint64(0x8))) & 0x7FF
						eid2 = uint64(ecs2.EntityID(uint32(eid2)).Index())
						e := p.ECS.Entities[uint32(eid2)]
						if e == nil {
							continue
						}
						if resolved != e {
							continue
						}
						parsed.ResolvedVictimPosition = &pv[len(pv)-1]
						break
					}
				}
			}
		case 4:
			err = binary.Read(r, binary.LittleEndian, &parsed.KillerUid)
			parsed.KillerUid &= 0x7FF
			if err != nil {
				return nil, err
			}
			if parsed.KillerUid != 0x7FF {
				resolved, ok := p.ECS.GetEntityByUid(int32(parsed.KillerUid))
				if !ok {
					return nil, fmt.Errorf("failed to resolve killer uid %v", parsed.KillerUid)
				}
				parsed.ResolvedKiller = resolved
				if p.Paths != nil {
					for eid, pv := range p.Paths.Paths {
						eid2 := ((uint64(uint64(eid)&0xff) << uint64(0x16)) | (uint64(eid) >> uint64(0x8))) & 0x7FF
						eid2 = uint64(ecs2.EntityID(uint32(eid2)).Index())
						e := p.ECS.Entities[uint32(eid2)]
						if e == nil {
							continue
						}
						if resolved != e {
							continue
						}
						parsed.ResolvedKillerPosition = &pv[len(pv)-1]
						break
					}
				}
			}
		case 5:
			_, err = r.ReadBytes(4)
			if err != nil {
				return nil, err
			}
		case 6:
			_, err = r.ReadByte()
			if err != nil {
				return nil, err
			}
		case 7:
			_, err = r.ReadByte()
			if err != nil {
				return nil, err
			}
		case 8:
			_, err = r.ReadBits(1)
			if err != nil {
				return nil, err
			}
		case 9:
			_, err = r.ReadByte()
			if err != nil {
				return nil, err
			}
		case 0xa:
			parsed.PlayerWeapon, err = r.ReadLenStr()
			if err != nil {
				return nil, err
			}
		case 0xb:
			err = binary.Read(r, binary.LittleEndian, &parsed.VictimPid)
			if err != nil {
				return nil, err
			}
		case 0xc:
			parsed.DestroyedWeapon, err = r.ReadLenStr()
			if err != nil {
				return nil, err
			}
		default:
			serializer.SkipReadingField(index, r)
		}
		index += 1
	}
	if p.KeepKills {
		p.Kills = append(p.Kills, parsed)
	}
	return parsed, nil
}
