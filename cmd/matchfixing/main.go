package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"log"
	"maps"
	"math"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"

	"github.com/maxsupermanhd/wrpl-inspector-private/parsers/ecs2"
	"github.com/maxsupermanhd/wrpl-inspector-private/parsers/slot2"
	"github.com/maxsupermanhd/wrpl-inspector/v2/wrpl"
	"github.com/maxsupermanhd/wrpl-inspector/v2/wrpl/packet"
	packetaward "github.com/maxsupermanhd/wrpl-inspector/v2/wrpl/packet/parser/award"
)

var (
	flECSHashesJSONPath = flag.String("ecshashes", "../../ecshashes.json", "path to ecshashes.json file")
	parserECSHashes     *ecs2.ComponentHashMaps
)

func main() {
	flag.Parse()
	parserECSHashes = noerr(ecs2.ReadComponentHashMaps(bytes.NewReader(noerr(os.ReadFile(*flECSHashesJSONPath)))))

	totals := map[string]int{}

	for _, dirPath := range append(flag.Args(), `/home/max/.var/app/com.valvesoftware.Steam/.local/share/Steam/steamapps/common/War Thunder/Replays`) {
		for _, dirFile := range noerr(os.ReadDir(dirPath))[3:] {
			if dirFile.IsDir() {
				continue
			}
			if !strings.HasSuffix(dirFile.Name(), ".wrpl") {
				continue
			}

			replayPath := filepath.Join(dirPath, dirFile.Name())
			ret, err := processReplay(replayPath)
			if err != nil {
				log.Println("processing ", replayPath, err.Error())
				continue
			}

			// if ret.TeamWon == 0 {
			// 	totals["inconclusive"] = totals["inconclusive"] + 1
			// 	continue
			// }

			verdict := "wtf"
			nst1 := float64(ret.Team1GroundCraftCounts[1] + ret.Team1GroundCraftCounts[2])
			nst2 := float64(ret.Team2GroundCraftCounts[1] + ret.Team2GroundCraftCounts[2])
			if nst1 == 0 {
				nst1 = 1
			}
			// should win pos 1 neg 2
			ratio := nst2 - nst1
			if nst1+nst2 <= 2 || math.Abs(ratio) < 2 {
				verdict = "fair"
			} else {
				if ret.TeamWon == 0 {
					verdict = "inconclusive"
				} else if ret.TeamWon == 1 && ratio > 0 {
					verdict = "MATCHFIXING"
				} else if ret.TeamWon == 1 && ratio < 0 {
					verdict = "gg"
				} else if ret.TeamWon == 2 && ratio < 0 {
					verdict = "MATCHFIXING"
				} else if ret.TeamWon == 2 && ratio > 0 {
					verdict = "gg"
				}
			}

			totals[verdict] = totals[verdict] + 1

			log.Printf("%v %v %3.1f(%2v) %3.1f(%2v) %7.2v %v", ret.Session, ret.TeamWon,
				ret.Team1GroundCraftsWeighted, nst1, ret.Team2GroundCraftsWeighted, nst2, ratio, verdict)
		}
	}
	for _, k := range slices.Sorted(maps.Keys(totals)) {
		v := totals[k]
		log.Println(k, v)
	}
}

func processReplay(replayPath string) (*SessionLineups, error) {
	replayFile, err := os.ReadFile(replayPath)
	if err != nil {
		return nil, err
	}

	rpl, err := wrpl.OpenReplay(bytes.NewReader(replayFile), false, true, true)
	if err != nil {
		return nil, err
	}
	defer rpl.Close()

	resultsBlk, err := wrpl.ParseBlk(rpl.Results)
	if err != nil {
		return nil, err
	}

	rSlot := &slot2.PacketSlotParser{}
	rAward := &packetaward.PacketAwardParser{}
	parserErrors, err := packet.ParsePacketsStreamed(packet.NewPacketStreamReader(rpl.PacketStream), []packet.PacketParser{rSlot, rAward})
	if err != nil {
		return nil, err
	}
	if len(parserErrors) > 0 {
		var ret strings.Builder
		fmt.Fprintf(&ret, "parsers have errors: (%d)", len(parserErrors))
		for _, v := range parserErrors {
			ret.WriteString("\n" + v.Error())
		}
		return nil, errors.New(ret.String())
	}

	ret := &SessionLineups{
		Session:                rpl.Header.SessionID,
		Team1GroundCraftCounts: map[int]int{},
		Team2GroundCraftCounts: map[int]int{},
	}

	for _, a := range slices.Backward(rAward.Awards) {
		if a.AwardName == "hidden_win_streak" {
			p := rSlot.Players[a.Player]
			if p != nil {
				ret.TeamWon = int(p.Team)
				break
			}
		}
	}

	ret.Lineups, err = processSetups(resultsBlk)
	if err != nil {
		return nil, err
	}

	sort.Slice(ret.Lineups, func(i, j int) bool {
		if ret.Lineups[i].Team != ret.Lineups[j].Team {
			return ret.Lineups[i].Team < ret.Lineups[j].Team
		}
		return ret.Lineups[i].UserID < ret.Lineups[j].UserID
	})

	for _, l := range ret.Lineups {
		craftTypes := []string{}
		groundCount := 0
		groundCountWeighted := 0.0
		groundAddScore := 1.0
		for _, c := range l.Crafts {
			if slices.Contains([]string{"tank", "heavy_tank", "tank_destroyer", "SPAA"}, c.Type) {
				groundCount++
				groundCountWeighted += groundAddScore
				groundAddScore /= 2
			} else if !slices.Contains([]string{"assault", "helicopter", "bomber", "fighter"}, c.Type) {
				log.Println(c.Type)
			}
			craftTypes = append(craftTypes, c.Type)
		}
		for _, p := range rSlot.Players {
			if p == nil {
				continue
			}
			if p.Uid.Player_id == uint64(l.UserID) {
				switch p.Team {
				case 1:
					ret.Team1GroundCraftsWeighted += groundCountWeighted
					ret.Team1GroundCraftCounts[groundCount] = ret.Team1GroundCraftCounts[groundCount] + 1
				case 2:
					ret.Team2GroundCraftsWeighted += groundCountWeighted
					ret.Team2GroundCraftCounts[groundCount] = ret.Team2GroundCraftCounts[groundCount] + 1
				default:
					panic(p.Team)
				}
			}
		}
		// log.Printf("%-011v %-015v %v %v %v", l.UserID, l.Platform, l.Team, groundCount, craftTypes)
	}

	return ret, nil
}

type SessionLineups struct {
	Session                   uint64
	Lineups                   []Lineup
	TeamWon                   int
	Team1GroundCraftCounts    map[int]int
	Team1GroundCraftsWeighted float64
	Team2GroundCraftCounts    map[int]int
	Team2GroundCraftsWeighted float64
}

type Lineup struct {
	UserID   int64
	Team     int
	Crafts   []Craft
	Platform string
}

type Craft struct {
	MRank int
	Name  string
	Type  string
}

func processSetups(resultsBlk map[string]any) ([]Lineup, error) {
	d1, ok := resultsBlk["uiScriptsData"].(map[string]any)
	if !ok {
		panic("no uiScriptsData")
		// return nil, nil
	}
	d2, ok := d1["playersInfo"].(map[string]any)
	if !ok {
		panic("no playersInfo")
		// return nil, nil
	}
	ret := []Lineup{}
	for _, v := range d2 {
		player, ok := v.(map[string]any)
		if !ok {
			panic("no player")
			// continue
		}
		l := Lineup{}
		l.UserID, _ = player["id"].(int64)
		playerTeam, _ := player["team"].(int64)
		l.Team = int(playerTeam)
		l.Platform, _ = player["platform"].(string)

		crafts, ok := player["crafts_info"].(map[string]any)
		if !ok {
			panic("no crafts_info")
			// continue
		}
		for k2, v2 := range crafts {
			if k2 == "__array" {
				continue
			}
			craft, ok := v2.(map[string]any)
			if !ok {
				panic("craft not map")
				// continue
			}
			cr := Craft{}
			mrank, _ := craft["mrank"].(int64)
			cr.MRank = int(mrank)
			cr.Name, _ = craft["name"].(string)
			cr.Type, _ = craft["type"].(string)
			l.Crafts = append(l.Crafts, cr)
		}
		ret = append(ret, l)
	}

	return ret, nil
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
