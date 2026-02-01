package ecs2

import (
	"encoding/binary"
	"fmt"
	"math/bits"

	"github.com/maxsupermanhd/wrpl-inspector/v2/wrpl/danet"
)

const MAX_FIELDS_NUM = 32

type IdFieldSerializer32 struct {
	sizes    [MAX_FIELDS_NUM]uint32
	currWrSz uint8 // probably not needed
	currRdSz uint8
}

func readSize(from *danet.BitReader) (uint32, error) {
	hdrT, err := from.ReadBits(3)
	if err != nil {
		return 0, err
	}
	hdr := hdrT[0]
	switch hdr {
	case 1:
		return 1, nil
	case 2:
		return 8, nil
	case 3:
		return 16, nil
	case 4:
		return 32, nil
	case 5:
		return 64, nil
	case 6:
		return 96, nil
	case 7:
		return 128, nil
	}
	sizeInBits, err := from.ReadCompressed()
	if err != nil {
		return 0, err
	}
	return uint32(sizeInBits), nil
}

func (p *IdFieldSerializer32) ReadFieldsSizeAndFlag(from *danet.BitReader) (uint32, error) {
	var fields uint64 // actually uint32
	var offset uint16
	start := from.BitOffset
	if start&7 != 0 {
		return 0, fmt.Errorf("IdFieldSerializer32 not alligned to byte")
	}
	err := binary.Read(from, binary.LittleEndian, &offset)
	if err != nil {
		return 0, err
	}
	fields, err = from.ReadCompressed()
	if err != nil {
		return 0, err
	}
	startBody := from.BitOffset
	from.BitOffset = int(offset<<3) + start
	p.currRdSz = uint8(bits.OnesCount32(uint32(fields)))

	for i := 0; i < int(p.currRdSz); i++ {
		p.sizes[i], err = readSize(from)
		if err != nil {
			return 0, err
		}
	}
	from.BitOffset = startBody
	return uint32(fields), nil
}

func (p *IdFieldSerializer32) SkipReadingField(index uint8, from *danet.BitReader) {
	from.BitOffset = from.BitOffset + int(p.sizes[index])
}
