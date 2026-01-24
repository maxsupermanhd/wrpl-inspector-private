package ecsui2

import (
	"fmt"
	"maps"
	"slices"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/maxsupermanhd/wrpl-inspector/inspector"
	packetecs "github.com/maxsupermanhd/wrpl-inspector/wrpl/packet/parser/ecs"
)

type ECSUI struct {
	rpl *inspector.LoadedReplay
	ecs *packetecs.PacketECSParser
}

func NewECSUI(rpl *inspector.LoadedReplay, ecs *packetecs.PacketECSParser) *ECSUI {
	return &ECSUI{
		rpl: rpl,
		ecs: ecs,
	}
}

func (tab *ECSUI) Name() string {
	return "ECS"
}

func (tab *ECSUI) Init() {
}

func (tab *ECSUI) Run() {
	imgui.TextUnformatted(fmt.Sprintf("Template defs: %d", len(tab.ecs.TemplateDefs)))
	for _, k := range slices.Sorted(maps.Keys(tab.ecs.TemplateDefs)) {
		v := tab.ecs.TemplateDefs[k]
		imgui.TextUnformatted(fmt.Sprintf("%v - %v (%v components)", v.ID, v.Name, len(v.Components)))
		flags := imgui.TableFlagsBorders | imgui.TableFlagsResizable | imgui.TableFlagsSizingFixedFit | imgui.TableFlagsNoHostExtendX
		if imgui.BeginTableV("components", 4, flags, imgui.Vec2{}, 0) {
			imgui.TableSetupColumn("n")
			imgui.TableSetupColumn("id")
			imgui.TableSetupColumn("type")
			imgui.TableSetupColumn("name")
			imgui.TableHeadersRow()

			for i2, v2 := range v.Components {
				imgui.TableNextRow()
				imgui.TableSetColumnIndex(0)
				imgui.TextUnformatted(fmt.Sprintf("%v", i2))
				imgui.TableSetColumnIndex(1)
				imgui.TextUnformatted(fmt.Sprintf("%v", v2))
				compDef, ok := tab.ecs.ComponentDefs[v2]
				if !ok {
					imgui.TableSetColumnIndex(2)
					imgui.TextUnformatted("not")
					imgui.TableSetColumnIndex(3)
					imgui.TextUnformatted("found")
				} else {
					imgui.TableSetColumnIndex(2)
					imgui.TextUnformatted(fmt.Sprintf("%v", compDef.Type))
					imgui.TableSetColumnIndex(3)
					imgui.TextUnformatted(fmt.Sprintf("%v", compDef.Name))
				}
			}
			imgui.EndTable()
		}
	}
	imgui.TextUnformatted(fmt.Sprintf("Component defs: %d", len(tab.ecs.ComponentDefs)))
	flags := imgui.TableFlagsBorders | imgui.TableFlagsResizable | imgui.TableFlagsSizingFixedFit | imgui.TableFlagsNoHostExtendX
	if imgui.BeginTableV("compdefs", 2, flags, imgui.Vec2{}, 0) {
		imgui.TableSetupColumn("name")
		imgui.TableSetupColumn("type")
		imgui.TableHeadersRow()
		for _, k := range slices.Sorted(maps.Keys(tab.ecs.ComponentDefs)) {
			v := tab.ecs.ComponentDefs[k]
			imgui.TableNextRow()
			imgui.TableNextColumn()
			imgui.TextUnformatted(fmt.Sprint(v.Name))
			imgui.TableNextColumn()
			imgui.TextUnformatted(fmt.Sprint(v.Type))
		}
		imgui.EndTable()
	}
	imgui.TextUnformatted(fmt.Sprintf("Messages: %d", len(tab.ecs.Messages)))
}
