package carve

import (
	"encoding/json"
	"errors"
	"fmt"
	"main/game"
	"main/parsers/critical"
	"main/parsers/ecs2"
	"main/parsers/fm"
	"main/parsers/kills2"
	nextsegmentparser "main/parsers/nextsegment"
	"main/parsers/paths"
	"main/parsers/severe"
	"main/parsers/slot2"
	"maps"
	"slices"

	"github.com/maxsupermanhd/wrpl-inspector/v2/wrpl"
	"github.com/maxsupermanhd/wrpl-inspector/v2/wrpl/packet"
	packetaward "github.com/maxsupermanhd/wrpl-inspector/v2/wrpl/packet/parser/award"
)

type MissionDefinition struct {
	Level         string
	LevelSettings string
	BattleType    string
}

type CarvedReplay struct {
	SessionID   uint64
	TimeStarted uint64

	Mission    MissionDefinition
	Difficulty byte

	GameDuration  float64
	TeamWon       byte
	Players       []SessionPlayer
	Kills         []SessionKill
	Awards        []SessionAward
	DamageReports []SessionDamage

	Entities    []SessionEntity
	CarveErrors []error
}

type SessionAward struct {
	Time      uint32
	AwardName string
	PlayerID  uint64
}

type SessionEntity struct {
	PlayerID    uint64
	EntityIndex uint32
	ModelName   string
	Path        SpaceTimeEncodeSummary
}

type SpaceTimeEncodeSummary []game.SpaceTime

func (e SpaceTimeEncodeSummary) MarshalJSON() ([]byte, error) {
	s := []game.SpaceTime(e)
	if len(s) == 0 {
		return []byte("null"), nil
	}
	return json.Marshal(map[string]any{
		"Start":        s[0],
		"End":          s[len(s)-1],
		"SamplesCount": len(s),
	})
}

func (e *SpaceTimeEncodeSummary) UnmarshalJSON(data []byte) error {
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
	awards := &packetaward.PacketAwardParser{}
	ecsp := ecs2.NewPacketECSParser(ecsHashes)
	prp := paths.NewPositionRetainerParser()
	fmp := &fm.PacketFlightModelParser{KeepResults: true, ECS: &ecsp.Mgr}
	kills := &kills2.PacketKillParser{KeepKills: true, ECS: &ecsp.Mgr, PathsGround: prp, PathsAir: fmp}
	sltp := &slot2.PacketSlotParser{}
	dcp := &critical.CriticalDamageParser{KeepResults: true, ECS: &ecsp.Mgr}
	dsp := &severe.SevereDamageParser{KeepResults: true, ECS: &ecsp.Mgr}
	pm := packet.NewParserMatcher([]packet.PacketParser{nsp, prp, ecsp, sltp, kills, awards, fmp, dcp, dsp})
	ret := &CarvedReplay{
		SessionID:   readers[parts[0]].Header.SessionID,
		TimeStarted: uint64(readers[parts[0]].Header.StartTime),
		Mission: MissionDefinition{
			Level:         carveHeaderString(readers[parts[0]].Header.Raw_Level[:]),
			LevelSettings: carveHeaderString(readers[parts[0]].Header.Raw_LevelSettings[:]),
			BattleType:    carveHeaderString(readers[parts[0]].Header.Raw_BattleType[:]),
		},
		Difficulty: readers[parts[0]].Header.Difficulty,
	}
	for parti, part := range parts {
		r := packet.NewPacketStreamReader(readers[part].PacketStream)
		pk := &packet.Packet{}
		for {
			isEOF, err := r.ReadPacket(pk)
			if isEOF {
				break
			}
			if err != nil {
				ret.CarveErrors = append(ret.CarveErrors, fmt.Errorf("reading packet %d from part %d: %w", pk.Seq, part, err))
				break
			}
			errs := pm.MatchIgnoreData(pk)
			if len(errs) > 0 {
				for _, err := range errs {
					ret.CarveErrors = append(ret.CarveErrors, fmt.Errorf("parsing packet %d from part %d returned error: %q", pk.Seq, part, err))
				}
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
		ret.CarveErrors = append(ret.CarveErrors, fmt.Errorf("parsing results blk: %w", err))
	}
	ret.GameDuration = getMapStringAnyValue(results, float64(-1), "timePlayed")
	ret.Players, err = assemblePlayers(results, sltp.Players)
	if err != nil {
		ret.CarveErrors = append(ret.CarveErrors, fmt.Errorf("assembling players: %w", err))
	}
	for _, a := range slices.Backward(awards.Awards) {
		if a.AwardName == "hidden_win_streak" {
			ret.TeamWon = sltp.Players[a.Player].Team
			break
		}
	}
	for _, a := range awards.Awards {
		player := sltp.Players[a.Player]
		if player == nil {
			continue
		}
		ret.Awards = append(ret.Awards, SessionAward{
			Time:      a.CurrentTime,
			AwardName: a.AwardName,
			PlayerID:  uint64(player.UserID),
		})
	}
	ret.Entities, err = assembleEntities(&ecsp.Mgr, sltp.Players, prp, fmp)
	if err != nil {
		ret.CarveErrors = append(ret.CarveErrors, fmt.Errorf("assembling entities: %w", err))
	}
	ret.Kills, err = assembleKills(&ecsp.Mgr, sltp.Players, kills)
	if err != nil {
		ret.CarveErrors = append(ret.CarveErrors, fmt.Errorf("assembling kills: %w", err))
	}
	ret.DamageReports, err = assembleDamage(&ecsp.Mgr, sltp.Players, dcp, dsp)
	if err != nil {
		ret.CarveErrors = append(ret.CarveErrors, fmt.Errorf("assembling damage: %w", err))
	}
	return ret, nil
}
