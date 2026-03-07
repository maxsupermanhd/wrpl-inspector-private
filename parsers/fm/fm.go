package fm

import (
	"encoding/binary"
	"errors"
	"fmt"
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
	HasUID bool
	UID    uint64
	Data   *FMData
}

type FMData struct {
	Unk0 bool
	Unk1 bool
	// Unk2        bool
	// Unk3        uint32
	// Unk4        bool
	Unk5       *FMDataUnk5
	Unk10      uint64 // len of Unk11
	Unk11      []uint64
	PosX       float32
	PosY       float32
	PosZ       float32
	EulerBytes uint32
	Unk13      [7]byte
	// EnginesData []FMEngineData
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

func (p *PacketFlightModelParser) Parse2(pk *packet.Packet) (*FMUpdatePacket, error) {
	ret := &FMUpdatePacket{}
	r := danet.NewBitReader(pk.PacketPayload)
	defer func() {
		// ret.Rem, _ = io.ReadAll(r)
		slices.Reverse(ret.Entries)
	}()
	eid := uint64(0)
	for {
		var err error
		e := &FMEntry{}

		// do we have eid
		e.HasUID, err = r.ReadBool()
		if err != nil {
			return ret, fmt.Errorf("reading new entry eid present bit: %w", err)
		}
		if e.HasUID {
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
		e.UID = eid
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

		r.IgnoreBits(33)
		// ed.Unk2, err = r.ReadBool()
		// if err != nil {
		// 	return ret, fmt.Errorf("reading unk2: %w", err)
		// }
		// ed.Unk3, err = r.ReadU32LE()
		// if err != nil {
		// 	return ret, fmt.Errorf("reading unk3: %w", err)
		// }

		r.IgnoreBits(1)
		// ed.Unk4, err = r.ReadBool()
		// if err != nil {
		// 	return ret, fmt.Errorf("reading unk4: %w", err)
		// }
		err = ignoreUnk5(r)
		// ed.Unk5, err = readUnk5(r)
		if err != nil {
			return ret, fmt.Errorf("reading unk5: %w", err)
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

		ed.EulerBytes, err = r.ReadU32LE()
		if err != nil {
			return ret, fmt.Errorf("reading euler bytes: %w", err)
		}
		r.IgnoreBits(32)
		// ed.Unk12, err = r.ReadU32LE()
		// if err != nil {
		// 	return ret, fmt.Errorf("reading unk12: %w", err)
		// }

		r.AlignToByteBoundary()

		_, err = r.ReadBitsInto(7*8, ed.Unk13[:])
		if err != nil {
			return ret, fmt.Errorf("reading unk13: %w", err)
		}

		err = ignoreEngines(r)
		if err != nil {
			return ret, fmt.Errorf("reading engines: %w", err)
		}

		err = ignoreSensors(r)
		if err != nil {
			return ret, fmt.Errorf("reading sensors: %w", err)
		}

		err = ignoreTargets(r)
		if err != nil {
			return ret, fmt.Errorf("reading targets: %w", err)
		}

		unk14, err := r.ReadBit()
		if err != nil {
			return ret, fmt.Errorf("reading unk14: %w", err)
		}
		if unk14 {
			r.IgnoreBits(32)
		}

		unk15, err := r.ReadBit()
		if err != nil {
			return ret, fmt.Errorf("reading unk15: %w", err)
		}
		if unk15 {
			unk16, err := r.ReadU32LE()
			if err != nil {
				return ret, fmt.Errorf("reading unk16: %w", err)
			}
			r.IgnoreBits(int(unk16))
		}

	}
	return ret, nil
}

func ignoreTargets(r *danet.BitReader) error {
	targetsCount := [1]byte{}
	_, err := r.ReadBitsInto(4, targetsCount[:])
	if err != nil {
		return fmt.Errorf("reading count: %w", err)
	}
	if targetsCount[0] > 8 {
		return fmt.Errorf("targets count > 8 (got %d)", targetsCount[0])
	}
	for i := range targetsCount[0] {
		err = ignoreTarget(r)
		if err != nil {
			return fmt.Errorf("reading target %d: %w", i, err)
		}
	}
	return nil
}

func ignoreTarget(r *danet.BitReader) error {
	r.IgnoreBits(16)
	unk0, err := r.ReadBit()
	if err != nil {
		return fmt.Errorf("reading unk0: %w", err)
	}
	if unk0 {
		r.IgnoreBits(32 * 3)
	}
	unk1, err := r.ReadBit()
	if err != nil {
		return fmt.Errorf("reading unk1: %w", err)
	}
	if unk1 {
		r.IgnoreBits(32*3 + 16*3)
	} else {
		r.IgnoreBits(32 * 6)
	}
	r.IgnoreBits(32 + 1 + 8)
	unk2, err := r.ReadBit()
	if err != nil {
		return fmt.Errorf("reading unk2: %w", err)
	}
	if unk2 {
		r.IgnoreBits(8)
	}
	r.IgnoreBits(1)
	unk3, err := r.ReadBit()
	if err != nil {
		return fmt.Errorf("reading unk3: %w", err)
	}
	r.IgnoreBits(1)
	if unk3 {
		r.IgnoreBits(8)
	}
	return nil
}

func ignoreSensors(r *danet.BitReader) error {
	sensorCount, err := r.ReadByte()
	if err != nil {
		return fmt.Errorf("reading sensor count: %w", err)
	}
	if sensorCount > 4 {
		return fmt.Errorf("sensor count > 4 (got %d)", sensorCount)
	}
	for i := range sensorCount {
		err = ignoreSensor(r)
		if err != nil {
			return fmt.Errorf("reading sensor %d: %w", i, err)
		}
	}
	if sensorCount > 0 {
		r.IgnoreBits(8)
	}
	return nil
}

func ignoreSensor(r *danet.BitReader) error {
	firstBool, err := r.ReadBit()
	if err != nil {
		return fmt.Errorf("reading first bool: %w", err)
	}
	sensorType, err := r.ReadByte()
	if err != nil {
		return fmt.Errorf("reading first bool: %w", err)
	}
	sensorType >>= 4
	switch sensorType {
	case 1:
		unk0, err := r.ReadBit()
		if err != nil {
			return fmt.Errorf("reading sensor type 1 unk0: %w", err)
		}
		if !unk0 {
			return nil
		}
		unk1, err := r.ReadU16LE()
		if err != nil {
			return fmt.Errorf("reading sensor type 1 unk1: %w", err)
		}
		r.IgnoreBits(32 + 16*3)
		if int16(unk1) < 0 {
			r.IgnoreBits(8)
		}
		unk2, err := r.ReadBit()
		if err != nil {
			return fmt.Errorf("reading sensor type 1 unk2: %w", err)
		}
		if unk2 {
			r.IgnoreBits(8)
		}
	case 2:
		if !firstBool {
			return nil
		}
		r.IgnoreBits(1 + 16 + 96 + 96 + 32)
	case 3:
		return errors.New("sensor type 3")
	case 4:
		unk0, err := r.ReadBit()
		if err != nil {
			return fmt.Errorf("reading sensor type 4 unk0: %w", err)
		}
		if unk0 {
			r.IgnoreBits(96)
		}
	}
	unk0, err := r.ReadBit()
	if err != nil {
		return fmt.Errorf("reading unk0: %w", err)
	}
	if unk0 {
		unk1 := [1]byte{}
		_, err := r.ReadBitsInto(6, unk1[:])
		if err != nil {
			return fmt.Errorf("reading unk1: %w", err)
		}
		r.IgnoreBits(32 * int(unk1[0]))
		r.IgnoreBits(6)
	}
	return nil
}

func ignoreEngines(r *danet.BitReader) error {
	enginesNum, err := r.ReadByte()
	if err != nil {
		return fmt.Errorf("reading engines num: %w", err)
	}
	if enginesNum > 0xf {
		return fmt.Errorf("engines num > 0xf (got %d)", enginesNum)
	}
	for range enginesNum {
		hasPower, err := r.ReadBit()
		if err != nil {
			return fmt.Errorf("reading has engine power: %w", err)
		}
		r.IgnoreBits(8)
		if hasPower {
			r.IgnoreBits(16)
		}
		unk2Present, err := r.ReadBit()
		if err != nil {
			return fmt.Errorf("reading unk2 present: %w", err)
		}
		if unk2Present {
			r.IgnoreBits(8)
		}
		r.IgnoreBits(16)
	}
	return nil
}

func readUnk5(r *danet.BitReader) (*FMDataUnk5, error) {
	unk5, err := r.ReadBool()
	if err != nil {
		return nil, fmt.Errorf("reading unk5: %w", err)
	}
	if unk5 {
		ret := &FMDataUnk5{}
		ret.Unk6, err = r.ReadBool()
		if err != nil {
			return ret, fmt.Errorf("reading unk6: %w", err)
		}
		ret.Unk7, err = r.ReadBool()
		if err != nil {
			return ret, fmt.Errorf("reading unk7: %w", err)
		}
		ret.Unk8, err = r.ReadBool()
		if err != nil {
			return ret, fmt.Errorf("reading unk8: %w", err)
		}
		bitsetLen := [1]byte{}
		_, err := r.ReadBitsInto(4, bitsetLen[:])
		if err != nil {
			return ret, fmt.Errorf("reading unk5 -> bitsetLen: %w", err)
		}
		ret.Unk9 = []bool{}
		for i := range bitsetLen[0] {
			bit, err := r.ReadBool()
			if err != nil {
				return ret, fmt.Errorf("reading unk5 -> bitset val %d/%d: %w", i, bitsetLen[0], err)
			}
			ret.Unk9 = append(ret.Unk9, bit)
		}
	}
	return nil, nil
}

func ignoreUnk5(r *danet.BitReader) error {
	unk5, err := r.ReadBool()
	if err != nil {
		return fmt.Errorf("ignoring unk5: %w", err)
	}
	if unk5 {
		r.IgnoreBits(3)
		bitsetLen := [1]byte{}
		_, err := r.ReadBitsInto(4, bitsetLen[:])
		if err != nil {
			return fmt.Errorf("reading unk5 -> bitsetLen: %w", err)
		}
		r.IgnoreBits(int(bitsetLen[0]))
	}
	return nil
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
