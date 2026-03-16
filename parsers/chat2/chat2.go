package chat2

import (
	"github.com/maxsupermanhd/wrpl-inspector/v2/wrpl/danet"
	"github.com/maxsupermanhd/wrpl-inspector/v2/wrpl/packet"
)

type ParsedPacketChatMessage struct {
	PacketSeq   uint64
	CurrentTime uint32
	Sender      string
	Content     string
	ChannelType byte
}

type PacketChatParser struct {
	Messages []ParsedPacketChatMessage
}

func (p *PacketChatParser) Name() string {
	return "chat2"
}

func (p *PacketChatParser) ParsesMatching() map[byte][][]packet.ParsingCondition {
	return map[byte][][]packet.ParsingCondition{
		3: nil,
	}
}

func (p *PacketChatParser) Parse(pk *packet.Packet) (any, error) {
	r := danet.NewBitReader(pk.PacketPayload)
	parsed := ParsedPacketChatMessage{
		PacketSeq:   pk.Seq,
		CurrentTime: pk.CurrentTime,
	}
	var err error
	parsed.Sender, err = r.ReadLenStr()
	if err != nil {
		return nil, err
	}
	parsed.Content, err = r.ReadLenStr()
	if err != nil {
		return nil, err
	}
	parsed.ChannelType, err = r.ReadByte()
	if err != nil {
		return nil, err
	}
	p.Messages = append(p.Messages, parsed)
	return &parsed, nil
}
