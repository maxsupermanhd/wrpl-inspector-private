package idfieldserializer

import (
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/maxsupermanhd/wrpl-inspector/v2/wrpl/danet"
)

func DeserializeIdFieldSerializer255(from *danet.BitReader, fieldReader func(fieldIndex uint16, fieldSize uint32) error) error {
	start := from.BitOffset
	if start&7 != 0 {
		return errors.New("IdFieldSerializer255 not alligned to byte")
	}

	var offset uint16
	err := binary.Read(from, binary.LittleEndian, &offset)
	if err != nil {
		return fmt.Errorf("reading offset: %w", err)
	}
	var count uint16
	err = binary.Read(from, binary.LittleEndian, &count)
	if err != nil {
		return fmt.Errorf("reading count: %w", err)
	}

	fieldsCount := count & ((1 << 12) - 1)
	if fieldsCount >= 255 {
		return fmt.Errorf("IdFieldSerializer255 fieldsCount >= 255 with %d (count is %d)", fieldsCount, count)
	}
	bitsPerID := count >> 12
	bitsForIndices := fieldsCount * bitsPerID

	startBody := from.BitOffset

	from.BitOffset = int(offset<<3) + start
	sizes := make([]uint32, fieldsCount)
	for i := range fieldsCount {
		sizes[i], err = ReadSize(from)
		if err != nil {
			return fmt.Errorf("reading size %d/%d: %w", i, fieldsCount, err)
		}
	}

	end := ((from.BitOffset + 7) >> 3) << 3
	_ = end

	indexes := make([]uint16, fieldsCount)
	from.BitOffset = start + int((offset<<3)-bitsForIndices)
	scratch := [2]byte{}
	for i := range fieldsCount {
		_, err = from.ReadBitsInto(int(bitsPerID), scratch[:])
		if err != nil {
			return fmt.Errorf("reading index %d/%d: %w", i, fieldsCount, err)
		}
		indexes[i] = binary.LittleEndian.Uint16(scratch[:])
		scratch[0] = 0
		scratch[1] = 0
	}

	from.BitOffset = startBody

	for i := range fieldsCount {
		s := sizes[i]
		idx := indexes[i]
		before := from.BitOffset
		err = fieldReader(idx, s)
		if err == ErrSkipField {
			from.BitOffset += int(s)
		} else if err != nil {
			return err
		}
		after := from.BitOffset
		if after-before != int(s) {
			return fmt.Errorf("field %d reader wrong len read %d but should %d", idx, after-before, s)
		}
	}
	return nil
}
