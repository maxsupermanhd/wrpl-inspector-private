package fm

import (
	"bytes"
	"fmt"
	"io"
	"main/parsers/ecs2"
	"slices"

	"github.com/maxsupermanhd/wrpl-inspector/v2/wrpl/danet"
	"github.com/maxsupermanhd/wrpl-inspector/v2/wrpl/packet"
)

type PacketFlightModelParser struct {
	ECS *ecs2.EntityManager
}

func (p *PacketFlightModelParser) Name() string {
	return "fm"
}

func (p *PacketFlightModelParser) ParsesMatching() map[byte][][]packet.ParsingCondition {
	return map[byte][][]packet.ParsingCondition{
		2: nil,
	}
}

type FMUpdatePacket struct {
	Rem     []byte
	Entries []*FMEntry
}

type FMEntry struct {
	HasEID     bool
	EID        uint64
	WasSkipped bool
	Unk0       [4]byte
	Unk1       byte
	Unk2       []byte
}

func (p *PacketFlightModelParser) Parse(pk *packet.Packet) (any, error) {
	ret := &FMUpdatePacket{}
	r := danet.NewBitReader(pk.PacketPayload)
	defer func() { // for clarity in inspector parsing results dump
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
		e.WasSkipped, err = r.ReadBool()
		if err != nil {
			return ret, fmt.Errorf("reading skip entry bit: %w", err)
		}
		if e.WasSkipped {
			continue
		}

		// at this point it's anyone's guess pretty much

		n, err := r.Read(e.Unk0[:])
		if err != nil {
			return ret, fmt.Errorf("reading unk0: %w", err)
		}
		if n != len(e.Unk0) {
			return ret, fmt.Errorf("reading skip entry bit: %w", err)
		}
		if !bytes.Equal(e.Unk0[:], []byte{0, 0, 0, 0}) {
			return ret, fmt.Errorf("unk0 is not full zero, don't know what to do now: %#v", e.Unk0)
		}

		e.Unk1, err = r.ReadByte()
		if err != nil {
			return ret, fmt.Errorf("reading unk1: %w", err)
		}
		if e.Unk1 != 0 {
			e.Unk2, err = r.ReadBits(9)
			if err != nil {
				return ret, fmt.Errorf("reading unk2: %w", err)
			}
		}

		return ret, fmt.Errorf("now what lol")
	}
	return ret, nil
}
