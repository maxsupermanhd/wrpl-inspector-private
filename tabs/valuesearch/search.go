package valuesearch

import (
	"bytes"
	"io"

	"github.com/klauspost/compress/zstd"
	"github.com/maxsupermanhd/wrpl-inspector/v2/inspector"
	"github.com/maxsupermanhd/wrpl-inspector/v2/inspector/imui"
	"github.com/maxsupermanhd/wrpl-inspector/v2/inspector/packetui"
	"github.com/maxsupermanhd/wrpl-inspector/v2/wrpl/packet"
)

type ValueSearchTab struct {
	rpl *inspector.LoadedReplay

	s               *inspector.PacketStreamSelector
	StreamProviders []packet.PacketStreamProvider

	ShouldUpdate bool

	view packetui.PacketStreamView
}

func NewValueSearchTab(rpl *inspector.LoadedReplay, providers ...packet.PacketStreamProvider) *ValueSearchTab {
	return &ValueSearchTab{
		rpl:             rpl,
		StreamProviders: providers,
	}
}

func (tab *ValueSearchTab) Name() string {
	return "Value search"
}

func (tab *ValueSearchTab) Init() {
	tab.s = inspector.NewPacketStreamSelector(append([]packet.PacketStreamProvider{tab.rpl.GlobalStreamProvider()}, tab.StreamProviders...)...)
	tab.update()
}

func (tab *ValueSearchTab) Run() {
	imui.FlagUpdate(&tab.ShouldUpdate, tab.s.Show())

	tab.view.Run()

	if tab.ShouldUpdate {
		tab.update()
	}
}

func (tab *ValueSearchTab) update() {
	tab.view.Stream = tab.view.Stream[:0]

	for _, pk := range tab.s.Stream() {
		for _, f := range findCompressions(pk.PacketPayload) {
			tab.view.Stream = append(tab.view.Stream, packet.ParsedPacket{
				Packet: packet.Packet{
					Seq:           pk.Seq,
					CurrentTime:   pk.CurrentTime,
					PacketType:    pk.PacketType,
					PacketPayload: f.Data,
				},
				ParsersResults: append(pk.ParsersResults, packet.ParserResult{
					Parser: "decompressed",
					Err:    f.Err,
					Data: DecompressedData{
						Offset:      f.Offset,
						Compression: f.Compression,
					},
				}),
			})
		}
	}
}

type DecompressedData struct {
	Offset      int
	Compression string
	Data        []byte
	Err         error
}

func findCompressions(payload []byte) (ret []DecompressedData) {
	i := bytes.Index(payload, []byte{0x28, 0xB5, 0x2F, 0xFD})
	if i != -1 {
		r, err := zstd.NewReader(bytes.NewReader(payload[i:]))
		if err != nil {
			ret = append(ret, DecompressedData{
				Offset:      i,
				Compression: "zstd",
				Data:        nil,
				Err:         err,
			})
		} else {
			data, err := io.ReadAll(r)
			if err != nil {
				ret = append(ret, DecompressedData{
					Offset:      i,
					Compression: "zstd",
					Data:        nil,
					Err:         err,
				})
			} else {
				ret = append(ret, DecompressedData{
					Offset:      i,
					Compression: "zstd",
					Data:        data,
				})
			}
		}
	}
	return
}
