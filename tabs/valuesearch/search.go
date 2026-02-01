package valuesearch

import (
	"bytes"

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
	tab.view.Stream = tab.s.Stream()
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
		if bytes.Contains(pk.PacketPayload, []byte{0x69, 0x42}) {
			tab.view.Stream = append(tab.view.Stream, pk)
		}
	}
}
