package fm

import (
	"encoding/binary"
	"fmt"
	"io"
	"main/parsers/ecs2"
	"math"
	"slices"

	"github.com/maxsupermanhd/wrpl-inspector/v2/wrpl/danet"
	"github.com/maxsupermanhd/wrpl-inspector/v2/wrpl/packet"
)

type PacketFlightModelParser struct {
	ECS         *ecs2.EntityManager
	KeepResults bool
	Results     []packet.ParsedPacket
}

func (p *PacketFlightModelParser) Name() string {
	return "fm"
}

func (p *PacketFlightModelParser) ParsesMatching() map[byte][][]packet.ParsingCondition {
	return map[byte][][]packet.ParsingCondition{
		2: nil,
	}
}

func (p *PacketFlightModelParser) GetPacketStreams() []packet.ParsedPacketStream {
	return []packet.ParsedPacketStream{{
		Name:    "rem",
		Packets: p.Results,
	}}
}

type FMUpdatePacket struct {
	Rem     []byte
	Entries []*FMEntry
}

type FMEntry struct {
	HasEID bool
	EID    uint64
	Data   *FMData
}

type FMData struct {
	Unk0       bool
	Unk1       bool
	Unk2       bool
	Unk3       uint32
	Unk4       bool
	Unk5       *FMDataUnk5
	Unk10      uint64
	Unk11      []uint64
	PosX       float32
	PosY       float32
	PosZ       float32
	EulerBytes []byte
	Unk12      []byte
}

type FMDataUnk5 struct {
	Unk6 bool
	Unk7 bool
	Unk8 bool
	Unk9 []bool
}

func (p *PacketFlightModelParser) Parse(pk *packet.Packet) (any, error) {
	ret, err := p.Parse2(pk)
	if p.KeepResults {
		p.Results = append(p.Results, packet.ParsedPacket{
			Packet: packet.Packet{
				Seq:           pk.Seq,
				CurrentTime:   pk.CurrentTime,
				PacketType:    2,
				PacketPayload: ret.Rem,
			},
			ParsersResults: []packet.ParserResult{{
				Parser: "fm",
				Data:   ret,
				Err:    err,
			}},
		})
	}
	return ret, err
}

// func (p *PacketFlightModelParser) Parse2(pk *packet.Packet) (any, error) {
func (p *PacketFlightModelParser) Parse2(pk *packet.Packet) (*FMUpdatePacket, error) {
	ret := &FMUpdatePacket{}
	r := danet.NewBitReader(pk.PacketPayload)
	defer func() {
		ret.Rem, _ = io.ReadAll(r)
		slices.Reverse(ret.Entries)
	}()
	eid := uint64(0)
	for {
		var err error
		e := &FMEntry{}

		// do we have eid
		e.HasEID, err = r.ReadBool()
		if err != nil {
			return ret, fmt.Errorf("reading new entry eid present bit: %w", err)
		}
		if e.HasEID {
			eid, err = r.ReadCompressed()
			if err != nil {
				return ret, fmt.Errorf("reading new eid: %w", err)
			}
		} else {
			eid++
		}

		// is this the end
		if eid == 16383 {
			break
		}
		e.EID = eid
		ret.Entries = append(ret.Entries, e)

		// does it have data
		noData, err := r.ReadBool()
		if err != nil {
			return ret, fmt.Errorf("reading skip entry bit: %w", err)
		}
		if noData {
			continue
		}
		ed := &FMData{}
		e.Data = ed

		ed.Unk0, err = r.ReadBool()
		if err != nil {
			return ret, fmt.Errorf("reading unk0: %w", err)
		}
		ed.Unk1, err = r.ReadBool()
		if err != nil {
			return ret, fmt.Errorf("reading unk1: %w", err)
		}
		if ed.Unk0 && ed.Unk1 {
			continue
		}

		ed.Unk2, err = r.ReadBool()
		if err != nil {
			return ret, fmt.Errorf("reading unk2: %w", err)
		}
		unk3b, err := r.ReadBytes(4)
		if err != nil {
			return ret, fmt.Errorf("reading unk3: %w", err)
		}
		ed.Unk3 = binary.LittleEndian.Uint32(unk3b)

		ed.Unk4, err = r.ReadBool()
		if err != nil {
			return ret, fmt.Errorf("reading unk4: %w", err)
		}
		unk5, err := r.ReadBool()
		if err != nil {
			return ret, fmt.Errorf("reading unk5: %w", err)
		}
		if unk5 {
			ed.Unk5 = &FMDataUnk5{}
			ed.Unk5.Unk6, err = r.ReadBool()
			if err != nil {
				return ret, fmt.Errorf("reading unk6: %w", err)
			}
			ed.Unk5.Unk7, err = r.ReadBool()
			if err != nil {
				return ret, fmt.Errorf("reading unk7: %w", err)
			}
			ed.Unk5.Unk8, err = r.ReadBool()
			if err != nil {
				return ret, fmt.Errorf("reading unk8: %w", err)
			}
			bitsetLen, err := r.ReadBits(4)
			if err != nil {
				return ret, fmt.Errorf("reading unk5 -> bitsetLen: %w", err)
			}
			ed.Unk5.Unk9 = []bool{}
			for i := range bitsetLen {
				bit, err := r.ReadBool()
				if err != nil {
					return ret, fmt.Errorf("reading unk5 -> bitset val %d/%d: %w", i, ed.Unk10, err)
				}
				ed.Unk5.Unk9 = append(ed.Unk5.Unk9, bit)
			}
		}

		ed.Unk10, err = r.ReadCompressed()
		if err != nil {
			return ret, fmt.Errorf("reading unk10: %w", err)
		}
		ed.Unk10 = -(ed.Unk10 & 1) ^ ed.Unk10>>1
		for i := range ed.Unk10 {
			val, err := r.ReadCompressed()
			if err != nil {
				return ret, fmt.Errorf("reading unk11 bitset val %d/%d: %w", i, ed.Unk10, err)
			}
			ed.Unk11 = append(ed.Unk11, val)
		}

		posXb, err := r.ReadBytes(4)
		if err != nil {
			return ret, fmt.Errorf("reading posX bytes: %w", err)
		}
		ed.PosX = math.Float32frombits(binary.LittleEndian.Uint32(posXb))
		posYb, err := r.ReadBytes(4)
		if err != nil {
			return ret, fmt.Errorf("reading posY bytes: %w", err)
		}
		ed.PosY = math.Float32frombits(binary.LittleEndian.Uint32(posYb))
		posZb, err := r.ReadBytes(4)
		if err != nil {
			return ret, fmt.Errorf("reading posZ bytes: %w", err)
		}
		ed.PosZ = math.Float32frombits(binary.LittleEndian.Uint32(posZb))

		ed.EulerBytes, err = r.ReadBytes(4)
		if err != nil {
			return ret, fmt.Errorf("reading euler bytes: %w", err)
		}

		r.AlignToByteBoundary()

		ed.Unk12, err = r.ReadBytes(8)
		if err != nil {
			return ret, fmt.Errorf("reading unk12: %w", err)
		}

		// return ret, fmt.Errorf("now what lol")
	}
	return ret, nil
}

func unpackEuler(packed uint32) (heading, attitude, bank float32) {
	heading = float32((packed>>19)&((1<<10)-1)) * float32(3.1415926535) / float32((1<<10)-1)
	attitude = float32((packed>>10)&((1<<9)-1)) * float32(1.570796326794895) / float32((1<<9)-1)
	bank = float32(packed&((1<<10)-1)) * float32(3.1415926535) / float32((1<<10)-1)
	if packed&(1<<31) != 0 {
		heading = -heading
	}
	if packed&(1<<30) != 0 {
		attitude = -attitude
	}
	if packed&(1<<29) != 0 {
		bank = -bank
	}
	return
}
