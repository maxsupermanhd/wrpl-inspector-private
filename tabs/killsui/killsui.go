package killsui

import (
	"fmt"
	"main/parsers/ecs2"
	"main/parsers/kills2"
	"strconv"
	"time"

	"github.com/AllenDang/cimgui-go/imgui"
	packetslot "github.com/maxsupermanhd/wrpl-inspector/v2/wrpl/packet/parser/slot"
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
	if imgui.BeginTableV("kills", 12, flags, imgui.Vec2{}, 0) {
		imgui.TableSetupColumn("n")
		imgui.TableSetupColumn("seq")
		imgui.TableSetupColumn("time")
		imgui.TableSetupColumn("KillerPid")
		imgui.TableSetupColumn("KillerUid")
		imgui.TableSetupColumn("KillerVehicle")
		imgui.TableSetupColumn("KillerWeapon")
		imgui.TableSetupColumn("VictimPid")
		imgui.TableSetupColumn("VictimUid")
		imgui.TableSetupColumn("DestroyedWeapon")
		imgui.TableSetupColumn("VehicleFromKillerUid")
		imgui.TableSetupColumn("VehicleFromVictimUid")

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
			imgui.TextUnformatted(fmt.Sprint(v.KillerPid))
			imgui.TableNextColumn()
			imgui.TextUnformatted(fmt.Sprint(v.KillerUid))
			imgui.TableNextColumn()
			imgui.TextUnformatted(fmt.Sprint(v.PlayerVehicle))
			imgui.TableNextColumn()
			imgui.TextUnformatted(fmt.Sprint(v.PlayerWeapon))
			imgui.TableNextColumn()
			imgui.TextUnformatted(fmt.Sprint(v.VictimPid))
			imgui.TableNextColumn()
			imgui.TextUnformatted(fmt.Sprint(v.VictimUid))
			imgui.TableNextColumn()
			imgui.TextUnformatted(fmt.Sprint(v.DestroyedWeapon))
			imgui.TableNextColumn()

			if v.ResolvedKiller == nil {
				imgui.TextUnformatted("!!unresolved killer!!")
			} else {
				name, _ := ecs2.GetObjectData[string](&v.ResolvedKiller.Data, "unit__className")
				imgui.TextUnformatted(fmt.Sprint(name))
			}
			imgui.TableNextColumn()

			if v.ResolvedVictim == nil {
				imgui.TextUnformatted("!!unresolved victim!!")
			} else {
				name, _ := ecs2.GetObjectData[string](&v.ResolvedVictim.Data, "unit__className")
				imgui.TextUnformatted(fmt.Sprint(name))
			}
		}
		imgui.EndTable()
	}
}
