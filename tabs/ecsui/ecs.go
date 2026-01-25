package ecsui2

import (
	"fmt"
	"main/parsers/ecs2"
	"maps"
	"slices"
	"strconv"
	"strings"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/davecgh/go-spew/spew"
	"github.com/maxsupermanhd/wrpl-inspector/inspector"
	"github.com/maxsupermanhd/wrpl-inspector/inspector/imui"
)

type ECSUI struct {
	rpl            *inspector.LoadedReplay
	ecs            *ecs2.PacketECSParser
	hashName       map[uint32]string
	hashType       map[uint32]string
	templateFilter string
}

func NewECSUI(rpl *inspector.LoadedReplay, ecs *ecs2.PacketECSParser, hashName, hashType map[uint32]string) *ECSUI {
	ret := &ECSUI{
		rpl:      rpl,
		ecs:      ecs,
		hashName: hashName,
		hashType: hashType,
	}
	if hashName == nil {
		hashName = map[uint32]string{}
	}
	if hashType == nil {
		hashType = map[uint32]string{}
	}
	return ret
}

func (tab *ECSUI) Name() string {
	return "ECS"
}

func (tab *ECSUI) Init() {
}

func (tab *ECSUI) Run() {
	imgui.InputTextWithHint("filter templates", "", &tab.templateFilter, 0, imui.ImEmptyInputCallback)
	if imgui.BeginTabBar("ECS Tabs") {
		if imgui.BeginTabItem("Components") {
			if imgui.BeginChildStr("components content") {
				tab.RunTabComponents()
				imgui.EndChild()
			}
			imgui.EndTabItem()
		}
		if imgui.BeginTabItem("Templates") {
			if imgui.BeginChildStr("templates content") {
				tab.RunTabTemplates()
				imgui.EndChild()
			}
			imgui.EndTabItem()
		}
		if imgui.BeginTabItem("Entities (by index)") {
			if imgui.BeginChildStr("entities index content") {
				tab.RunTabEntities()
				imgui.EndChild()
			}
			imgui.EndTabItem()
		}
		if imgui.BeginTabItem("Entities (by eid)") {
			if imgui.BeginChildStr("entities eid content") {
				tab.RunTabEntitiesUidLookup()
				imgui.EndChild()
			}
			imgui.EndTabItem()
		}
		imgui.EndTabBar()
	}
}

func (tab *ECSUI) RunTabEntities() {
	for _, entIdx := range slices.Sorted(maps.Keys(tab.ecs.Mgr.Entities)) {
		ent := tab.ecs.Mgr.Entities[entIdx]
		if !strings.Contains(ent.Template, tab.templateFilter) {
			continue
		}
		imgui.TextUnformatted(fmt.Sprintf("Entity %v %q", entIdx, ent.Template))
		flags := imgui.TableFlagsBorders | imgui.TableFlagsResizable | imgui.TableFlagsSizingFixedFit | imgui.TableFlagsNoHostExtendX
		if imgui.BeginTableV("compdefs", 4, flags, imgui.Vec2{}, 0) {
			imgui.TableSetupColumn("n")
			imgui.TableSetupColumn("name")
			imgui.TableSetupColumn("type hash")
			imgui.TableSetupColumn("value")
			imgui.TableHeadersRow()
			for i, k := range slices.Sorted(maps.Keys(ent.Data.Components)) {
				v := ent.Data.Components[k]
				imgui.TableNextRow()
				imgui.TableNextColumn()
				imgui.TextUnformatted(strconv.FormatInt(int64(i), 10))
				imgui.TableNextColumn()
				imgui.TextUnformatted(k)
				imgui.TableNextColumn()
				imgui.TextUnformatted(fmt.Sprintf("0x%08X", v.Type))
				imgui.TableNextColumn()
				imgui.TextUnformatted(spew.Sdump(v.Value))
			}
			imgui.EndTable()
		}
	}
}

func (tab *ECSUI) RunTabEntitiesUidLookup() {
	for _, uid := range slices.Sorted(maps.Keys(tab.ecs.Mgr.Uid_lookup)) {
		ent := tab.ecs.Mgr.Uid_lookup[uid]
		if !strings.Contains(ent.Template, tab.templateFilter) {
			continue
		}
		if imgui.TreeNodeStr(fmt.Sprintf("Entity %v %q", uid, ent.Template)) {
			flags := imgui.TableFlagsBorders | imgui.TableFlagsResizable | imgui.TableFlagsSizingFixedFit | imgui.TableFlagsNoHostExtendX
			if imgui.BeginTableV("compdefs", 4, flags, imgui.Vec2{}, 0) {
				imgui.TableSetupColumn("n")
				imgui.TableSetupColumn("name")
				imgui.TableSetupColumn("type hash")
				imgui.TableSetupColumn("value")
				imgui.TableHeadersRow()
				for i, k := range slices.Sorted(maps.Keys(ent.Data.Components)) {
					v := ent.Data.Components[k]
					imgui.TableNextRow()
					imgui.TableNextColumn()
					imgui.TextUnformatted(strconv.FormatInt(int64(i), 10))
					imgui.TableNextColumn()
					imgui.TextUnformatted(k)
					imgui.TableNextColumn()
					imgui.TextUnformatted(fmt.Sprintf("0x%08X", v.Type))
					imgui.TableNextColumn()
					imgui.TextUnformatted(spew.Sdump(v.Value))
				}
				imgui.EndTable()
			}
			imgui.TreePop()
		}
	}
}

func (tab *ECSUI) RunTabComponents() {
	imgui.TextUnformatted(fmt.Sprintf("Component defs: %d", len(tab.ecs.ComponentDefs)))
	flags := imgui.TableFlagsBorders | imgui.TableFlagsResizable | imgui.TableFlagsSizingFixedFit | imgui.TableFlagsNoHostExtendX
	if imgui.BeginTableV("compdefs", 5, flags, imgui.Vec2{}, 0) {
		imgui.TableSetupColumn("n")
		imgui.TableSetupColumn("name hash")
		imgui.TableSetupColumn("name")
		imgui.TableSetupColumn("type hash")
		imgui.TableSetupColumn("type")
		imgui.TableHeadersRow()
		for i, k := range slices.Sorted(maps.Keys(tab.ecs.ComponentDefs)) {
			v := tab.ecs.ComponentDefs[k]
			imgui.TableNextRow()
			imgui.TableNextColumn()
			imgui.TextUnformatted(strconv.FormatInt(int64(i), 10))
			imgui.TableNextColumn()
			imgui.TextUnformatted(fmt.Sprintf("0x%08X", v.Name))
			imgui.TableNextColumn()
			imgui.TextUnformatted(tab.hashName[uint32(v.Name)])
			imgui.TableNextColumn()
			imgui.TextUnformatted(fmt.Sprintf("0x%08X", v.Type))
			imgui.TableNextColumn()
			imgui.TextUnformatted(tab.hashType[uint32(v.Type)])
		}
		imgui.EndTable()
	}
}

func (tab *ECSUI) RunTabTemplates() {
	imgui.TextUnformatted(fmt.Sprintf("Template defs: %d", len(tab.ecs.TemplateDefs)))
	for _, k := range slices.Sorted(maps.Keys(tab.ecs.TemplateDefs)) {
		v := tab.ecs.TemplateDefs[k]
		if !strings.Contains(v.Name, tab.templateFilter) {
			continue
		}
		imgui.TextUnformatted(fmt.Sprintf("%v - %v (%v components)", v.ID, v.Name, len(v.Components)))
		flags := imgui.TableFlagsBorders | imgui.TableFlagsResizable | imgui.TableFlagsSizingFixedFit | imgui.TableFlagsNoHostExtendX
		if imgui.BeginTableV("components", 6, flags, imgui.Vec2{}, 0) {
			imgui.TableSetupColumn("n")
			imgui.TableSetupColumn("idx")
			imgui.TableSetupColumn("type hash")
			imgui.TableSetupColumn("type")
			imgui.TableSetupColumn("name hash")
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
					imgui.TableSetColumnIndex(4)
					imgui.TextUnformatted("found")
				} else {
					imgui.TableSetColumnIndex(2)
					imgui.TextUnformatted(fmt.Sprintf("0x%08X", compDef.Type))
					imgui.TableSetColumnIndex(3)
					imgui.TextUnformatted(fmt.Sprintf("%v", tab.hashType[uint32(compDef.Type)]))
					imgui.TableSetColumnIndex(4)
					imgui.TextUnformatted(fmt.Sprintf("0x%08X", compDef.Name))
					imgui.TableSetColumnIndex(5)
					imgui.TextUnformatted(fmt.Sprintf("%v", tab.hashName[uint32(compDef.Name)]))
				}
			}
			imgui.EndTable()
		}
	}
}
