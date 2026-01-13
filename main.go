package main

import (
	"encoding/json"
	"main/parsers/stub0"
	"main/parsers/stub1"
	"main/tabs/interpreter2"
	"main/tabs/resultsui"
	"main/tabs/valuesearch"
	"os"

	"github.com/AllenDang/cimgui-go/backend/glfwbackend"
	"github.com/davecgh/go-spew/spew"
	"github.com/maxsupermanhd/wrpl-inspector/inspector"
	basictabs "github.com/maxsupermanhd/wrpl-inspector/inspector/basicTabs"
	"github.com/maxsupermanhd/wrpl-inspector/inspector/ecsui"
	"github.com/maxsupermanhd/wrpl-inspector/inspector/packetui"
	"github.com/maxsupermanhd/wrpl-inspector/inspector/playersui"
	"github.com/maxsupermanhd/wrpl-inspector/wrpl"
	"github.com/maxsupermanhd/wrpl-inspector/wrpl/packet"
	packetaward "github.com/maxsupermanhd/wrpl-inspector/wrpl/packet/parser/award"
	packetchat "github.com/maxsupermanhd/wrpl-inspector/wrpl/packet/parser/chat"
	packetecs "github.com/maxsupermanhd/wrpl-inspector/wrpl/packet/parser/ecs"
	packetkill "github.com/maxsupermanhd/wrpl-inspector/wrpl/packet/parser/kill"
	packetmovement "github.com/maxsupermanhd/wrpl-inspector/wrpl/packet/parser/movement"
	packetslot "github.com/maxsupermanhd/wrpl-inspector/wrpl/packet/parser/slot"
)

// ^0f8bfe0e090001(........)
func main() {
	ui := &inspector.UI{
		InitFont:        noerr(os.ReadFile("HackNerdFontMono-Regular.ttf")),
		ProcessReplayFn: replayProcessor,
		InitWindowFlags: map[glfwbackend.GLFWWindowFlags]int{
			glfwbackend.GLFWWindowFlagsMaximized: 1,
		},
	}
	ui.Run()
}

func replayProcessor(rpl *inspector.LoadedReplay) ([]packet.PacketParser, []inspector.Tab) {
	ecs := packetecs.NewPacketECSParser()
	slot := &packetslot.PacketSlotParser{KeepMessages: true}
	parsers := []packet.PacketParser{
		&packetchat.PacketChatParser{},
		&packetaward.PacketAwardParser{},
		&packetkill.PacketKillParser{},
		&packetmovement.PacketMovementParser{},
		&stub0.PacketStubParser{},
		&stub1.PacketStubParser{},
		ecs,
		slot,
	}
	streams := []packet.PacketStreamProvider{ecs, slot}
	tabs := []inspector.Tab{}
	tabs = append(tabs, basictabs.NewBasicSummaryTab(rpl))
	tabs = append(tabs, basictabs.NewBasicTextTab("Header", spew.Sdump(rpl.Header)))
	tabs = append(tabs, genBlkJSONTab("Settings raw", rpl.Settings))
	tabs = append(tabs, genBlkJSONTab("Results raw", rpl.Results))
	tabs = append(tabs, noerr(resultsui.NewResultsTab(rpl.Results)))
	tabs = append(tabs, packetui.NewPacketsTab(rpl, streams...))
	tabs = append(tabs, ecsui.NewECSUI(rpl, ecs))
	tabs = append(tabs, interpreter2.NewByteInterpreterTab(rpl, streams...))
	tabs = append(tabs, valuesearch.NewValueSearchTab(rpl, streams...))
	tabs = append(tabs, playersui.NewPlayersUI(rpl, slot))
	return parsers, tabs
}

func genBlkJSONTab(name string, data []byte) inspector.Tab {
	if len(data) == 0 {
		return basictabs.NewBasicTextTab(name, "no data")
	}
	dataAny, err := wrpl.ParseBlk(data)
	if err != nil {
		dataAny = map[string]any{
			"blk parse error": err,
		}
	}
	dataJSON, err := json.MarshalIndent(dataAny, "", "\t")
	if err != nil {
		return basictabs.NewBasicTextTab(name, "Marshal error: "+err.Error())
	}
	return basictabs.NewBasicTextTab(name, string(dataJSON))
	// return basictabs.NewBasicTextTab(name, spew.Sdump(dataAny))
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func noerr[T any](ret T, err error) T {
	must(err)
	return ret
}

func noerr2[T, T2 any](ret T, ret2 T2, err error) (T, T2) {
	must(err)
	return ret, ret2
}
