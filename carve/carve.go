package carve

import (
	"bytes"
	"errors"
	"fmt"
	"main/parsers/ecs2"
	"main/parsers/fm"
	"main/parsers/kills2"
	nextsegmentparser "main/parsers/nextsegment"
	"main/parsers/paths"
	"main/parsers/slot2"
	"maps"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/maxsupermanhd/wrpl-inspector/v2/wrpl"
	"github.com/maxsupermanhd/wrpl-inspector/v2/wrpl/packet"
	packetaward "github.com/maxsupermanhd/wrpl-inspector/v2/wrpl/packet/parser/award"
	"github.com/rs/zerolog/log"
)

type MissionDefinition struct {
	Level         string
	LevelSettings string
	BattleType    string
}

type CarvedReplay struct {
	SessionID   uint64
	Mission     MissionDefinition
	Difficulty  byte
	TimeStarted time.Time
	TimePlayed  float64
	TeamWon     byte
	Players     []SessionPlayer
	Kills       []SessionKill
	Awards      []SessionAward
	Entities    []SessionEntity
	CarveErrors []error
}

type SessionPlayer struct {
	UserID  uint64
	Name    string
	ClanTag string
	Team    byte

	Crafts map[string]string

	SessionPlayerPerformance
}

type SessionPlayerPerformance struct {
	SquadID             int  `json:"squadId"`
	AutoSquad           bool `json:"autoSquad"`
	Deaths              int  `json:"deaths"`
	AiGroundKills       int  `json:"aiGroundKills"`
	AiKills             int  `json:"aiKills"`
	AiNavalKills        int  `json:"aiNavalKills"`
	Assists             int  `json:"assists"`
	AwardDamage         int  `json:"awardDamage"`
	CaptureZone         int  `json:"captureZone"`
	DamageZone          int  `json:"damageZone"`
	GroundKills         int  `json:"groundKills"`
	HumanKills          int  `json:"humanKills"`
	Kills               int  `json:"kills"`
	MissileEvades       int  `json:"missileEvades"`
	NavalKills          int  `json:"navalKills"`
	Score               int  `json:"score"`
	ShellIntgerceptions int  `json:"shellIntgerceptions"`
	TeamKills           int  `json:"teamKills"`
}

type SessionKill struct {
	Time           uint32
	KillerID       uint64
	KillerModel    string
	KillerPosition *paths.SpaceTime
	Weapon         string
	VictimID       uint64
	VictimModel    string
	VictimPosition *paths.SpaceTime
}

type SessionAward struct {
	Time      uint32
	AwardName string
	PlayerID  uint64
}

type SessionEntity struct {
	PlayerID  uint64
	ModelName string
	Path      EncodedSpaceTime
}

type EncodedSpaceTime []paths.SpaceTime

func (e EncodedSpaceTime) MarshalJSON() ([]byte, error) {
	return []byte("0"), nil
}

func (e *EncodedSpaceTime) UnmarshalJSON(data []byte) error {
	return nil
}

func CarveReplay(readers map[int]*wrpl.ReplayReader, ecsHashes ecs2.ComponentHashMaps) (*CarvedReplay, error) {
	if len(readers) == 0 {
		return nil, errors.New("no replays?")
	}
	parts := slices.Collect(maps.Keys(readers))
	slices.Sort(parts)
	sid := readers[parts[0]].Header.SessionID
	if !slices.Contains(parts, 0) {
		return nil, fmt.Errorf("carving session %d has no part 0", sid)
	}
	for _, v := range parts {
		if readers[v].Header.SessionID != sid {
			return nil, fmt.Errorf("multiple sessions: %d vs %d", v, readers[v].Header.SessionID)
		}
	}
	err := carveCheckPartsContinuity(parts)
	if err != nil {
		return nil, fmt.Errorf("discontinuity for session %d: %w", sid, err)
	}
	nsp := &nextsegmentparser.PacketNextSegmentParser{}
	prp := paths.NewPositionRetainerParser()
	sltp := &slot2.PacketSlotParser{}
	ecsp := ecs2.NewPacketECSParser(ecsHashes)
	kills := &kills2.PacketKillParser{KeepKills: true, ECS: &ecsp.Mgr, Paths: prp}
	awards := &packetaward.PacketAwardParser{}
	fmp := &fm.PacketFlightModelParser{KeepResults: true, ECS: &ecsp.Mgr}
	pm := packet.NewParserMatcher([]packet.PacketParser{nsp, prp, ecsp, sltp, kills, awards, fmp})
	for parti, part := range parts {
		r := packet.NewPacketStreamReader(readers[part].PacketStream)
		pk := &packet.Packet{}
		for {
			isEOF, err := r.ReadPacket(pk)
			if isEOF {
				break
			}
			if err != nil {
				return nil, fmt.Errorf("reading packet %d from part %d (sid %d): %w", pk.Seq, part, sid, err)
			}
			errs := pm.MatchIgnoreData(pk)
			if len(errs) > 0 {
				errsStr := []string{}
				for _, err := range errs {
					errsStr = append(errsStr, err.Error())
				}
				return nil, fmt.Errorf("parsing packet %d from part %d returned errors: %q (sid %d)", pk.Seq, part, strings.Join(errsStr, ", "), sid)
			}
			pk.Seq++
		}
		isLast := parti == len(parts)-1
		if isLast {
			if nsp.LastSeq == pk.Seq-1 {
				return nil, fmt.Errorf("part %d is last but has next segment packet (got parts %#+v) (sid %d)", part, parts, sid)
			}
		} else {
			if nsp.LastSeq != pk.Seq-1 {
				return nil, fmt.Errorf("part %d does not have next segment packet but we have more parts (got parts %#+v) (sid %d)", part, parts, sid)
			}
		}
	}
	results, err := wrpl.ParseBlk(readers[parts[len(parts)-1]].Results)
	if err != nil {
		return nil, fmt.Errorf("parsing results blk: %w", err)
	}
	ret := &CarvedReplay{
		SessionID:   readers[parts[0]].Header.SessionID,
		TimeStarted: time.Unix(int64(readers[parts[0]].Header.StartTime), 0),
		TimePlayed:  getMapStringAnyValue(results, float64(-1), "timePlayed"),
		Mission: MissionDefinition{
			Level:         carveHeaderString(readers[parts[0]].Header.Raw_Level[:]),
			LevelSettings: carveHeaderString(readers[parts[0]].Header.Raw_LevelSettings[:]),
			BattleType:    carveHeaderString(readers[parts[0]].Header.Raw_BattleType[:]),
		},
		Difficulty: readers[parts[0]].Header.Difficulty,
	}
	for _, a := range slices.Backward(awards.Awards) {
		if a.AwardName == "hidden_win_streak" {
			ret.TeamWon = sltp.Players[a.Player].Team
			break
		}
	}
	resultsPlayers, _ := results["player"].([]any)
	resultsMatchingInfo, _ := results["matchingInfo"].(map[string]any)
	for _, p := range sltp.Players {
		if p == nil {
			continue
		}
		if p.Name == "" {
			continue
		}
		if p.UserID == 0 {
			continue
		}
		player := SessionPlayer{
			UserID:  uint64(p.UserID),
			Name:    p.Name,
			ClanTag: p.ClanTag,
			Team:    p.Team,
		}
		for _, r := range resultsPlayers {
			r2, ok := r.(map[string]any)
			if !ok {
				continue
			}
			pidStr, ok := r2["userId"].(string)
			if !ok {
				continue
			}
			pid, err := strconv.ParseUint(pidStr, 10, 64)
			if err != nil {
				continue
			}
			if p.UserID != uint32(pid) {
				continue
			}

			player.SessionPlayerPerformance.SquadID = int(getMapStringAnyValue(r2, int64(-1), "squadID"))
			player.SessionPlayerPerformance.AutoSquad = getMapStringAnyValue(r2, false, "autoSquad")
			player.SessionPlayerPerformance.Deaths = int(getMapStringAnyValue(r2, int64(-1), "deaths"))
			player.SessionPlayerPerformance.AiGroundKills = int(getMapStringAnyValue(r2, int64(-1), "aiGroundKills"))
			player.SessionPlayerPerformance.AiKills = int(getMapStringAnyValue(r2, int64(-1), "aiKills"))
			player.SessionPlayerPerformance.AiNavalKills = int(getMapStringAnyValue(r2, int64(-1), "aiNavalKills"))
			player.SessionPlayerPerformance.Assists = int(getMapStringAnyValue(r2, int64(-1), "assists"))
			player.SessionPlayerPerformance.AwardDamage = int(getMapStringAnyValue(r2, int64(-1), "awardDamage"))
			player.SessionPlayerPerformance.CaptureZone = int(getMapStringAnyValue(r2, int64(-1), "captureZone"))
			player.SessionPlayerPerformance.DamageZone = int(getMapStringAnyValue(r2, int64(-1), "damageZone"))
			player.SessionPlayerPerformance.GroundKills = int(getMapStringAnyValue(r2, int64(-1), "groundKills"))
			player.SessionPlayerPerformance.HumanKills = int(getMapStringAnyValue(r2, int64(-1), "humanKills"))
			player.SessionPlayerPerformance.Kills = int(getMapStringAnyValue(r2, int64(-1), "kills"))
			player.SessionPlayerPerformance.MissileEvades = int(getMapStringAnyValue(r2, int64(-1), "missileEvades"))
			player.SessionPlayerPerformance.NavalKills = int(getMapStringAnyValue(r2, int64(-1), "navalKills"))
			player.SessionPlayerPerformance.Score = int(getMapStringAnyValue(r2, int64(-1), "score"))
			player.SessionPlayerPerformance.ShellIntgerceptions = int(getMapStringAnyValue(r2, int64(-1), "shellIntgerceptions"))
			player.SessionPlayerPerformance.TeamKills = int(getMapStringAnyValue(r2, int64(-1), "teamKills"))
		}
		for rid, r := range resultsMatchingInfo {
			r2, ok := r.(map[string]any)
			if !ok {
				continue
			}
			pid, err := strconv.ParseUint(rid, 10, 64)
			if err != nil {
				continue
			}
			if p.UserID != uint32(pid) {
				continue
			}
			craftsInfo, ok := r2["crafts_info"].(map[string]any)
			if !ok {
				continue
			}
			player.Crafts = map[string]string{}
			for _, k := range slices.Sorted(maps.Keys(craftsInfo)) {
				craft, ok := craftsInfo[k].(map[string]any)
				if !ok {
					continue
				}
				k2 := strings.TrimPrefix(k, "array")
				_, err := strconv.Atoi(k2)
				if err != nil {
					continue
				}
				craftName, ok := craft["name"].(string)
				if !ok {
					continue
				}
				player.Crafts[k2] = craftName
			}
		}
		ret.Players = append(ret.Players, player)
	}
	for eid, movement := range prp.Paths {
		eid2 := ((uint64(uint64(eid)&0xff) << uint64(0x16)) | (uint64(eid) >> uint64(0x8))) & 0x7FF
		eid2 = uint64(ecs2.EntityID(uint32(eid2)).Index())
		e := ecsp.Mgr.Entities[uint32(eid2)]
		if e == nil {
			continue
		}
		playerid, ok := ecs2.GetObjectData[int32](&e.Data, "unit__playerId")
		if !ok {
			continue
		}
		if playerid < 0 || playerid >= 255 {
			continue
		}
		player := sltp.Players[playerid]
		if player == nil {
			continue
		}
		modelName, ok := ecs2.GetObjectData[string](&e.Data, "unit__className")
		if !ok {
			continue
		}
		ret.Entities = append(ret.Entities, SessionEntity{
			PlayerID:  uint64(player.UserID),
			ModelName: modelName,
			Path:      movement,
		})
	}
	flyingEntities := map[*ecs2.Entity]SessionEntity{}
	for _, e0 := range fmp.Results {
		entry := e0.ParsersResults[0].Data.(*fm.FMEntry)
		if entry.Data == nil {
			continue
		}
		if entry.ResolvedEntity == nil {
			continue
		}
		entity, ok := flyingEntities[entry.ResolvedEntity]
		if !ok {
			playerid, ok := ecs2.GetObjectData[int32](&entry.ResolvedEntity.Data, "unit__playerId")
			if !ok {
				continue
			}
			if playerid < 0 || playerid >= 255 {
				continue
			}
			player := sltp.Players[playerid]
			if player == nil {
				continue
			}
			modelName, ok := ecs2.GetObjectData[string](&entry.ResolvedEntity.Data, "unit__className")
			if !ok {
				continue
			}
			entity = SessionEntity{
				PlayerID:  uint64(player.UserID),
				ModelName: modelName,
			}
		}
		entity.Path = append(entity.Path, paths.SpaceTime{
			Time: e0.CurrentTime,
			X:    int64(entry.Data.PosX),
			Y:    int64(entry.Data.PosY),
			Z:    int64(entry.Data.PosZ),
		})
		flyingEntities[entry.ResolvedEntity] = entity
	}
	ret.Entities = append(ret.Entities, slices.Collect(maps.Values(flyingEntities))...)
	for _, k := range kills.Kills {
		if k.ResolvedKiller == nil || k.ResolvedVictim == nil {
			continue
		}
		if k.ResolvedVictimPosition == nil {
			continue
		}
		entry := SessionKill{
			Time:           k.CurrentTime,
			Weapon:         k.PlayerWeapon,
			KillerPosition: k.ResolvedKillerPosition,
			VictimPosition: k.ResolvedVictimPosition,
		}
		if k.ResolvedKillerPosition != nil {
			entry.KillerPosition = k.ResolvedKillerPosition
		}
		entry.KillerModel, _ = ecs2.GetObjectData[string](&k.ResolvedKiller.Data, "unit__className")
		killerID, _ := ecs2.GetObjectData[int32](&k.ResolvedKiller.Data, "unit__playerId")
		if killerID < 0 || killerID >= int32(len(sltp.Players)) {
			log.Warn().Msgf("oob unit__playerId??? %d", killerID)
			continue
		} else {
			p := sltp.Players[killerID]
			if p == nil {
				log.Warn().Msgf("nil player %d", killerID)
				continue
			} else {
				entry.KillerID = uint64(p.UserID)
			}
		}
		entry.VictimModel, _ = ecs2.GetObjectData[string](&k.ResolvedVictim.Data, "unit__className")
		victimID, _ := ecs2.GetObjectData[int32](&k.ResolvedVictim.Data, "unit__playerId")
		if victimID < 0 || victimID >= int32(len(sltp.Players)) {
			log.Warn().Msgf("oob unit__playerId??? %d", victimID)
			continue
		} else {
			p := sltp.Players[victimID]
			if p == nil {
				log.Warn().Msgf("nil player %d", victimID)
				continue
			} else {
				entry.VictimID = uint64(p.UserID)
			}
		}
		ret.Kills = append(ret.Kills, entry)
	}
	return ret, nil
}

func carveCheckPartsContinuity(parts []int) error {
	prevPart := -1
	// 0  1  3  5  7  9...
	for _, v := range parts {
		if v%2 == 1 {
			if prevPart+2 != v {
				return fmt.Errorf("found orderd part %d but previous was %d (got parts %#+v)", v, prevPart, parts)
			} else {
				prevPart = v
			}
		}
	}
	return nil
}

func carveHeaderString(s []byte) string {
	return string(bytes.Trim(s, "\x00"))
}

func SnailDifficultyToStr(difficulty byte) string {
	return []string{"arcade", "realistic", "hardcore"}[(difficulty>>2)&3]
}

func getMapStringAnyValue[T any](m map[string]any, d T, p ...string) T {
	if len(p) == 0 {
		return d
	}
	if len(p) == 1 {
		ret, ok := m[p[0]].(T)
		if ok {
			return ret
		} else {
			return d
		}
	}
	next, ok := m[p[0]].(map[string]any)
	if !ok {
		return d
	}
	return getMapStringAnyValue(next, d, p[1:]...)
}
