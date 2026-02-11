package fm

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"main/parsers/ecs2"
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
	Unk0     [4]byte
	Unk1     byte
	Unk2     []byte
	Unk3     byte
	Heading  float32
	Altitude float32
	Bank     float32
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

		// at this point it's anyone's guess pretty much

		n, err := r.Read(ed.Unk0[:])
		if err != nil {
			return ret, fmt.Errorf("reading unk0: %w", err)
		}
		if n != len(ed.Unk0) {
			return ret, fmt.Errorf("reading skip entry bit: %w", err)
		}
		if !bytes.Equal(ed.Unk0[:], []byte{0, 0, 0, 0}) && !bytes.Equal(ed.Unk0[:], []byte{0x20, 0, 0, 0}) {
			return ret, fmt.Errorf("unk0 is not full zero, don't know what to do now: %#v", ed.Unk0)
		}

		ed.Unk1, err = r.ReadByte()
		if err != nil {
			return ret, fmt.Errorf("reading unk1: %w", err)
		}
		if ed.Unk1 != 0 {
			ed.Unk2, err = r.ReadBits(9)
			if err != nil {
				return ret, fmt.Errorf("reading unk2: %w", err)
			}
		}

		ed.Unk3, err = r.ReadByte()
		if err != nil {
			return ret, fmt.Errorf("reading unk3: %w", err)
		}
		if ed.Unk3 != 0x10 {
			return ret, fmt.Errorf("unk3 is not 0x10, don't know what to do now: %#v", ed.Unk3)
		}

		r.IgnoreBits(5)
		r.IgnoreBytes(4)
		packedBytes, err := r.ReadBytes(4)
		if err != nil {
			return ret, fmt.Errorf("reading unk3: %w", err)
		}
		ed.Heading, ed.Altitude, ed.Bank = unpackEuler(binary.LittleEndian.Uint32(packedBytes))

		return ret, fmt.Errorf("now what lol")
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
