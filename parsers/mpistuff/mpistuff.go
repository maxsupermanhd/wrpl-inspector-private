package mpiparser

import (
	"encoding/binary"

	"github.com/maxsupermanhd/wrpl-inspector/v2/wrpl/danet"
	"github.com/maxsupermanhd/wrpl-inspector/v2/wrpl/packet"
)

type MPIStuffParser struct {
}

func (p *MPIStuffParser) Name() string {
	return "mpistuff"
}

func (p *MPIStuffParser) ParsesMatching() map[byte][][]packet.ParsingCondition {
	return map[byte][][]packet.ParsingCondition{
		4: nil,
	}
}

const (
	MPIInvalidObjectID = 0xFFFF
	MPIObjectExtID     = 0xFFFFFFFF
)

type MPIMessage struct {
	Oid       uint16
	Ext       uint32
	MessageID uint16
}

func (p *MPIStuffParser) Parse(pk *packet.Packet) (any, error) {
	r := danet.NewBitReader(pk.PacketPayload)
	ret := MPIMessage{
		Ext: MPIObjectExtID,
		Oid: MPIInvalidObjectID,
	}
	err := binary.Read(r, binary.LittleEndian, &ret.Oid)
	if err != nil {
		return nil, err
	}
	if ret.Oid&(1<<10) == 1 && ret.Oid != MPIInvalidObjectID {
		ret.Oid ^= 1 << 10
		ext, err := r.ReadCompressed()
		if err != nil {
			return nil, err
		}
		ret.Ext = uint32(ext)
	}
	err = binary.Read(r, binary.LittleEndian, &ret.MessageID)
	if err != nil {
		return nil, err
	}

	return ret, err
}
