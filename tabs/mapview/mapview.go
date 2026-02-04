package mapview

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"main/parsers/ecs2"
	"main/parsers/kills2"
	"main/parsers/paths"
	"maps"
	"math"
	"slices"
	"strconv"

	"github.com/AllenDang/cimgui-go/backend"
	"github.com/AllenDang/cimgui-go/backend/glfwbackend"
	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/davecgh/go-spew/spew"
	"github.com/maxsupermanhd/wrpl-inspector/v2/inspector"
	"github.com/maxsupermanhd/wrpl-inspector/v2/inspector/imui"
	packetslot "github.com/maxsupermanhd/wrpl-inspector/v2/wrpl/packet/parser/slot"
)

type MapViewTab struct {
	Backend backend.Backend[glfwbackend.GLFWWindowFlags]
	Rpl     *inspector.LoadedReplay
	Kills   *kills2.PacketKillParser
	Ecs     *ecs2.EntityManager
	Players *packetslot.PacketSlotParser
	Paths   *paths.PositionRetainerParser

	TankMapsPath string
	DataminePath string

	initErr         error
	tankmapTextureW int
	tankmapTextureH int
	tankmapTexture  *imgui.TextureRef
	rLevel          string
	rLevelSettings  string
	rBattleType     string
	rAreas          map[string]AreaDef
	rMission        MissionDef
	rOffsets        levelDef
	rMainAreaName   string
	rImageArea      image.Rectangle
	rCaps           map[int]AreaDef
	rCoordScaleX    float64
	rCoordScaleZ    float64

	highlightPath uint64

	imOutSize imgui.Vec2
	imOutSp   imgui.Vec2
}

func (tab *MapViewTab) Name() string {
	return "MapView"
}

func (tab *MapViewTab) Init() {
	if tab.Rpl == nil {
		tab.initErr = errors.New("rpl is nil")
		return
	}
	tab.rLevel = string(bytes.Trim(tab.Rpl.Header.Raw_Level[:], "\x00"))
	tab.rLevelSettings = string(bytes.Trim(tab.Rpl.Header.Raw_LevelSettings[:], "\x00"))
	tab.rBattleType = string(bytes.Trim(tab.Rpl.Header.Raw_BattleType[:], "\x00"))

	tankmapImage, err := levelToTankmap(tab.TankMapsPath, tab.rLevel)
	if err != nil {
		tab.initErr = fmt.Errorf("levelToTankmap: %w", err)
		return
	}
	tab.tankmapTextureW = tankmapImage.Rect.Dx()
	tab.tankmapTextureH = tankmapImage.Rect.Dy()
	tex := tab.Backend.CreateTextureRgba(tankmapImage, tankmapImage.Rect.Dx(), tankmapImage.Rect.Dy())
	tab.tankmapTexture = &tex

	mission, err := missionLoad(tab.DataminePath, tab.rLevelSettings)
	if err != nil {
		tab.initErr = fmt.Errorf("missionLoad: %w", err)
		return
	}
	tab.rMission = *mission

	areas := map[string]AreaDef{}
	if mission != nil && mission.Areas != nil {
		areas = mission.Areas
	}
	tab.rAreas = areas

	offsets, err := getLevelCoords(tab.DataminePath, tab.rLevel)
	if err != nil {
		tab.initErr = fmt.Errorf("getLevelCoords: %w", err)
		return
	}
	tab.rOffsets = *offsets

	mainArea, mainAreaName, err := findMainBattleArea(areas, tab.rBattleType, tab.Rpl.Header.Difficulty)
	if err != nil {
		tab.initErr = fmt.Errorf("findMainBattleArea: %w", err)
		return
	}
	tab.rMainAreaName = mainAreaName
	mainAreaX, mainAreaZ, mainAreaW, mainAreaH := calcDrawArea(offsets, mainArea, false)
	tab.rImageArea = image.Rect(int(mainAreaX), int(mainAreaZ), int(mainAreaX+mainAreaW), int(mainAreaZ+mainAreaH))

	caps, err := findCaps(areas, tab.rBattleType, tab.Rpl.Header.Difficulty)
	if err != nil {
		caps = map[int]AreaDef{}
	}
	tab.rCaps = caps

	tab.rCoordScaleX = math.Abs(offsets.TankMapCoord1[0]-offsets.TankMapCoord0[0]) / 2048
	tab.rCoordScaleZ = math.Abs(offsets.TankMapCoord1[1]-offsets.TankMapCoord0[1]) / 2048
}

func (tab *MapViewTab) Run() {
	if imgui.BeginChildStrV("map view tab content", imgui.ContentRegionAvail(), 0, imgui.WindowFlagsHorizontalScrollbar) {
		tab.RunContent()
	}
	imgui.EndChild()
}

func (tab *MapViewTab) RunContent() {
	avail := imgui.ContentRegionAvail()
	tab.imOutSize = imgui.Vec2{X: min(avail.X, avail.Y), Y: min(avail.X, avail.Y)}
	tab.imOutSp = imgui.CursorScreenPos()
	if tab.initErr != nil {
		imgui.PushTextWrapPos()
		imgui.TextUnformatted("init error: " + tab.initErr.Error())
		imgui.PopTextWrapPos()
		if imgui.Button("ignore") {
			tab.initErr = nil
		}
	} else {
		tab.DrawView()
	}

	imgui.Dummy(tab.imOutSize)
	imgui.SameLine()
	if imgui.BeginChildStr("map side view") {
		if imgui.BeginTabBar("side tabs") {
			if imgui.BeginTabItem("General") {
				tab.runGeneral()
				imgui.EndTabItem()
			}
			if imgui.BeginTabItem("Paths") {
				tab.runPaths()
				imgui.EndTabItem()
			}
			if imgui.BeginTabItem("Areas") {
				tab.runAreas()
				imgui.EndTabItem()
			}
			if imgui.BeginTabItem("Offsets") {
				tab.runOffsets()
				imgui.EndTabItem()
			}
			imgui.EndTabBar()
		}
	}
	imgui.EndChild()
}

func (tab *MapViewTab) DrawView() {
	dl := imgui.WindowDrawList()

	uv0 := imgui.Vec2{X: float32(tab.rImageArea.Min.X) / float32(tab.tankmapTextureW), Y: float32(tab.rImageArea.Min.Y) / float32(tab.tankmapTextureH)}
	uv1 := imgui.Vec2{X: float32(tab.rImageArea.Max.X) / float32(tab.tankmapTextureW), Y: float32(tab.rImageArea.Max.Y) / float32(tab.tankmapTextureH)}
	dlImage(dl, *tab.tankmapTexture, tab.imOutSp, tab.imOutSize, uv0, uv1)

	sw := float64(tab.imOutSize.X / float32(tab.rImageArea.Dx()))
	sh := float64(tab.imOutSize.Y / float32(tab.rImageArea.Dy()))
	for _, eid := range slices.Sorted(maps.Keys(tab.Paths.Paths)) {
		path := tab.Paths.Paths[eid]
		if eid == tab.highlightPath {
			continue
		}
		for _, pos := range path {
			x := (((float64(pos.X) - tab.rOffsets.TankMapCoord0[0]) / tab.rCoordScaleX) - float64(tab.rImageArea.Min.X)) * sw
			z := ((2048 - (float64(pos.Z)-tab.rOffsets.TankMapCoord0[1])/tab.rCoordScaleZ) - float64(tab.rImageArea.Min.Y)) * sh
			coords := imgui.Vec2{X: float32(x), Y: float32(z)}
			dl.PathLineTo(tab.imOutSp.Add(coords))
		}
		dl.PathStroke(0xFFFFFFFF)
	}
	hpath := tab.Paths.Paths[tab.highlightPath]
	if hpath != nil {
		for _, pos := range hpath {
			x := (((float64(pos.X) - tab.rOffsets.TankMapCoord0[0]) / tab.rCoordScaleX) - float64(tab.rImageArea.Min.X)) * sw
			z := ((2048 - (float64(pos.Z)-tab.rOffsets.TankMapCoord0[1])/tab.rCoordScaleZ) - float64(tab.rImageArea.Min.Y)) * sh
			coords := imgui.Vec2{X: float32(x), Y: float32(z)}
			dl.PathLineTo(tab.imOutSp.Add(coords))
		}
		dl.PathStrokeV(0xFF0000FF, 0, 3)
	}
}

func (tab *MapViewTab) runGeneral() {
	imgui.TextUnformatted("Level: " + tab.rLevel)
	imgui.TextUnformatted("Level settings: " + tab.rLevelSettings)
	imgui.TextUnformatted("Battle type: " + tab.rBattleType)
	imgui.TextUnformatted("Main area: " + tab.rMainAreaName)
	imgui.TextUnformatted(fmt.Sprint("Offsets: ", tab.rOffsets))
	imgui.TextUnformatted(fmt.Sprint("Image area: ", tab.rImageArea))
	imgui.TextUnformatted(fmt.Sprint("Image width: ", tab.rImageArea.Dx()))
	imgui.TextUnformatted(fmt.Sprint("Image height: ", tab.rImageArea.Dy()))
	imgui.TextUnformatted(fmt.Sprint("Coord scale X: ", tab.rCoordScaleX))
	imgui.TextUnformatted(fmt.Sprint("Coord scale Z: ", tab.rCoordScaleZ))
	imgui.TextUnformatted(fmt.Sprint("Sp: ", tab.imOutSp))
	imgui.TextUnformatted(fmt.Sprint("Out size: ", tab.imOutSize))
	imgui.TextUnformatted(fmt.Sprint("Aspect: ", float32(tab.rImageArea.Dx())/float32(tab.rImageArea.Dy())))
}

func (tab *MapViewTab) runPaths() {
	flags := imgui.TableFlagsBorders | imgui.TableFlagsResizable | imgui.TableFlagsSizingFixedFit | imgui.TableFlagsNoHostExtendX
	if imgui.BeginTableV("paths", 6, flags, imgui.Vec2{}, 0) {
		imgui.TableSetupColumn("eid dec")
		imgui.TableSetupColumn("eid hex")
		imgui.TableSetupColumn("eid2 dec")
		imgui.TableSetupColumn("eid2 hex")
		imgui.TableSetupColumn("samples")
		imgui.TableSetupColumn("player")
		imgui.TableHeadersRow()
		for _, eid := range slices.Sorted(maps.Keys(tab.Paths.Paths)) {
			path := tab.Paths.Paths[eid]
			// eid2 := ((uint64(uint64(eid)&0xff) << uint64(0x16)) | (uint64(eid) >> uint64(0x8))) & 0x7FF
			// eid2 := eid & 0x7FF
			eid2 := eid >> 8
			imgui.TableNextRow()
			if tab.highlightPath == eid {
				imgui.TableSetBgColor(imgui.TableBgTargetRowBg1, 0x22FFFFFF)
			}
			imgui.TableNextColumn()
			if imgui.Button(strconv.FormatUint(eid, 10)) {
				tab.highlightPath = eid
			}
			imgui.TableNextColumn()
			imgui.TextUnformatted(strconv.FormatUint(eid, 16))
			imgui.TableNextColumn()
			imgui.TextUnformatted(strconv.FormatUint(eid2, 10))
			imgui.TableNextColumn()
			imgui.TextUnformatted(strconv.FormatUint(eid2, 16))
			imgui.TableNextColumn()
			imgui.TextUnformatted(strconv.FormatInt(int64(len(path)), 10))
			imgui.TableNextColumn()
			e := tab.Ecs.Entities[uint32(eid2)]
			if e == nil {
				imgui.TextUnformatted("unresolved entity")
				continue
			}
			playerid, ok := ecs2.GetObjectData[int32](&e.Data, "unit__playerId")
			if !ok {
				imgui.TextUnformatted("unresolved unit__playerId")
				continue
			}
			if playerid < 0 || playerid >= 255 {
				imgui.TextUnformatted("unit__playerId oob")
				continue
			}
			player := tab.Players.Players[playerid]
			if player == nil {
				imgui.TextUnformatted("nil player")
				continue
			}
			imgui.TextUnformatted(player.Name)
		}
	}
	imgui.EndTable()
}

func (tab *MapViewTab) runAreas() {
	for _, k := range slices.Sorted(maps.Keys(tab.rAreas)) {
		v := tab.rAreas[k]
		imgui.TextUnformatted(fmt.Sprintf("%q %q", k, v.Type))
		for i2, v2 := range v.TM {
			imgui.TextUnformatted(fmt.Sprintf("%d %v", i2, v2))
		}
	}
}

func (tab *MapViewTab) runOffsets() {
	s := spew.Sprint(tab.rOffsets)
	imgui.InputTextMultiline("##basictext", &s, imgui.ContentRegionAvail(), imgui.InputTextFlagsReadOnly, imui.ImEmptyInputCallback)
}

func (tab *MapViewTab) Close() {
	if tab.tankmapTexture != nil {
		tab.Backend.DeleteTexture(*tab.tankmapTexture)
	}
}

func ImSliderVec2(label string, v *imgui.Vec2, min, max float32) bool {
	sh := [2]float32{v.X, v.Y}
	ret := imgui.SliderFloat2(label, &sh, min, max)
	if ret {
		v.X = sh[0]
		v.Y = sh[1]
	}
	return ret
}

func dlImage(dl *imgui.DrawList, ref imgui.TextureRef, sp, size, uv0, uv1 imgui.Vec2) {
	dl.AddImageQuadV(ref,
		sp, sp.Add(imgui.Vec2{X: size.X}), sp.Add(imgui.Vec2{X: size.X, Y: size.Y}), sp.Add(imgui.Vec2{Y: size.Y}),
		uv0, imgui.Vec2{X: uv1.X, Y: uv0.Y}, imgui.Vec2{X: uv1.X, Y: uv1.Y}, imgui.Vec2{X: uv0.X, Y: uv1.Y}, 0xFFFFFFFF)
}
