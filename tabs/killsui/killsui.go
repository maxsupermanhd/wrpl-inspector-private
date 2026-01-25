package killsui

import (
	"fmt"
	"main/parsers/ecs2"
	"main/parsers/kills2"
	"strconv"
	"time"

	"github.com/AllenDang/cimgui-go/imgui"
	packetslot "github.com/maxsupermanhd/wrpl-inspector/wrpl/packet/parser/slot"
)

type KillsTab struct {
	kills   *kills2.PacketKillParser
	ecs     *ecs2.EntityManager
	players *packetslot.PacketSlotParser
}

func NewKillsTab(kills *kills2.PacketKillParser, ecs *ecs2.EntityManager, players *packetslot.PacketSlotParser) *KillsTab {
	return &KillsTab{
		kills:   kills,
		ecs:     ecs,
		players: players,
	}
}

func (tab *KillsTab) Name() string {
	return "Kills2"
}

func (tab *KillsTab) Init() {

}

func (tab KillsTab) Run() {
	imgui.TextUnformatted(fmt.Sprintf("Kills: %d, ECS entities: %d", len(tab.kills.Kills), len(tab.ecs.Uid_lookup)))
	flags := imgui.TableFlagsBorders | imgui.TableFlagsResizable | imgui.TableFlagsSizingFixedFit | imgui.TableFlagsNoHostExtendX
	if imgui.BeginTableV("kills", 10, flags, imgui.Vec2{}, 0) {
		imgui.TableSetupColumn("n")
		imgui.TableSetupColumn("seq")
		imgui.TableSetupColumn("time")
		imgui.TableSetupColumn("control")
		imgui.TableSetupColumn("killerID")
		imgui.TableSetupColumn("killer vehicle")
		imgui.TableSetupColumn("killerUID")
		imgui.TableSetupColumn("victimUID")
		imgui.TableSetupColumn("weapon")
		imgui.TableSetupColumn("victimID")
		imgui.TableHeadersRow()
		for i, v := range tab.kills.Kills {
			imgui.TableNextRow()
			imgui.TableNextColumn()
			imgui.TextUnformatted(strconv.FormatInt(int64(i), 10))
			imgui.TableNextColumn()
			imgui.TextUnformatted(strconv.FormatInt(int64(v.Seq), 10))
			imgui.TableNextColumn()
			imgui.TextUnformatted((time.Duration(v.CurrentTime) * time.Millisecond).String())
			imgui.TableNextColumn()
			imgui.TextUnformatted(fmt.Sprint(v.Control))
			imgui.TableNextColumn()
			imgui.TextUnformatted(fmt.Sprint(v.KillerID))
			imgui.TableNextColumn()
			imgui.TextUnformatted(fmt.Sprint(v.KillerVehicle))
			imgui.TableNextColumn()
			imgui.TextUnformatted(fmt.Sprint(v.KillerUID))
			imgui.TableNextColumn()
			imgui.TextUnformatted(fmt.Sprint(v.VictimUID))
			imgui.TableNextColumn()
			imgui.TextUnformatted(fmt.Sprint(v.Weapon))
			imgui.TableNextColumn()
			imgui.TextUnformatted(fmt.Sprint(v.VictimID))
		}
		imgui.EndTable()
	}
}
