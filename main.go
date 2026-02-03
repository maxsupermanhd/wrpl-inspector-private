package main

import (
	"bytes"
	"encoding/json"
	"main/parsers/ecs2"
	"main/parsers/kills2"
	"main/parsers/stub0"
	"main/parsers/stub1"
	"main/tabs/bitshiftui"
	ecsui2 "main/tabs/ecsui"
	"main/tabs/interpreter2"
	"main/tabs/killsui"
	"main/tabs/resultsui"
	"main/tabs/valuesearch"
	"os"

	"github.com/AllenDang/cimgui-go/backend/glfwbackend"
	"github.com/davecgh/go-spew/spew"
	"github.com/maxsupermanhd/wrpl-inspector/v2/inspector"
	basictabs "github.com/maxsupermanhd/wrpl-inspector/v2/inspector/basicTabs"
	"github.com/maxsupermanhd/wrpl-inspector/v2/inspector/packetui"
	"github.com/maxsupermanhd/wrpl-inspector/v2/inspector/playersui"
	"github.com/maxsupermanhd/wrpl-inspector/v2/wrpl"
	"github.com/maxsupermanhd/wrpl-inspector/v2/wrpl/packet"
	packetaward "github.com/maxsupermanhd/wrpl-inspector/v2/wrpl/packet/parser/award"
	packetchat "github.com/maxsupermanhd/wrpl-inspector/v2/wrpl/packet/parser/chat"
	packetmovement "github.com/maxsupermanhd/wrpl-inspector/v2/wrpl/packet/parser/movement"
	packetslot "github.com/maxsupermanhd/wrpl-inspector/v2/wrpl/packet/parser/slot"
)

var (
	chms *ecs2.ComponentHashMaps
)

// ^0f8bfe0e090001(........)
func main() {
	chms = noerr(ecs2.ReadComponentHashMaps(bytes.NewReader(noerr(os.ReadFile("ecshashes.json")))))
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
	ecs := ecs2.NewPacketECSParser(*chms)
	slot := &packetslot.PacketSlotParser{KeepMessages: true}
	kills := &kills2.PacketKillParser{
		KeepKills: true,
		ECS:       &ecs.Mgr,
	}
	parsers := []packet.PacketParser{
		&packetchat.PacketChatParser{},
		&packetaward.PacketAwardParser{},
		kills,
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
	hashTypes := chms.ComponentNames
	hashNames := map[uint32]string{}
	for k, v := range chms.DataComponents {
		hashNames[k] = v.Name
	}
	tabs = append(tabs, ecsui2.NewECSUI(rpl, ecs, hashNames, hashTypes))
	tabs = append(tabs, interpreter2.NewByteInterpreterTab(rpl, streams...))
	tabs = append(tabs, bitshiftui.NewBitShiftUI())
	tabs = append(tabs, valuesearch.NewValueSearchTab(rpl, streams...))
	tabs = append(tabs, playersui.NewPlayersUI(rpl, slot))
	tabs = append(tabs, killsui.NewKillsTab(kills, &ecs.Mgr, slot))
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
