package carve

import (
	"wrplinspectorprivate/game"
	"wrplinspectorprivate/parsers/ecs2"
	"wrplinspectorprivate/parsers/kills2"
	"wrplinspectorprivate/parsers/slot2"
)

type SessionKill struct {
	Time              uint32
	KillerID          uint64
	KillerModel       string
	KillerEntityIndex uint32
	KillerPosition    *game.SpaceTime
	Weapon            string
	VictimID          uint64
	VictimModel       string
	VictimEntityIndex uint32
	VictimPosition    *game.SpaceTime
}

func assembleKills(ecs *ecs2.EntityManager, players [256]*slot2.Player, kills *kills2.PacketKillParser) (ret []SessionKill, err error) {
	for _, k := range kills.Kills {
		if k.ResolvedVictim == nil {
			continue
		}
		entry := SessionKill{
			Time:           k.CurrentTime,
			Weapon:         k.PlayerWeapon,
			KillerPosition: k.ResolvedKillerPosition,
			VictimPosition: k.ResolvedVictimPosition,
		}
		entry.KillerID, entry.KillerModel = resolveEntityDetails(players, k.ResolvedKiller)
		entry.KillerEntityIndex = resolveEntityToEntityIndex(ecs.Entities, k.ResolvedKiller)
		entry.VictimID, entry.VictimModel = resolveEntityDetails(players, k.ResolvedVictim)
		entry.VictimEntityIndex = resolveEntityToEntityIndex(ecs.Entities, k.ResolvedVictim)
		ret = append(ret, entry)
	}
	return
}
