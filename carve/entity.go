package carve

import (
	"maps"
	"slices"

	"github.com/maxsupermanhd/wrpl-inspector-private/game"
	"github.com/maxsupermanhd/wrpl-inspector-private/parsers/ecs2"
	"github.com/maxsupermanhd/wrpl-inspector-private/parsers/fm"
	"github.com/maxsupermanhd/wrpl-inspector-private/parsers/paths"
	"github.com/maxsupermanhd/wrpl-inspector-private/parsers/slot2"
)

func assembleEntities(
	ecs *ecs2.EntityManager,
	players [256]*slot2.Player,
	positionsGround *paths.PositionRetainerParser,
	positionsAir *fm.PacketFlightModelParser,
) (ret []SessionEntity, err error) {
	for eid, movement := range positionsGround.Paths {
		eid2 := ((uint64(uint64(eid)&0xff) << uint64(0x16)) | (uint64(eid) >> uint64(0x8))) & 0x7FF
		eid2 = uint64(ecs2.EntityID(uint32(eid2)).Index())
		e := ecs.Entities[uint32(eid2)]
		if e == nil {
			continue
		}
		entry := SessionEntity{
			Path: movement,
		}
		entry.PlayerID, entry.ModelName = resolveEntityDetails(players, e)
		entry.EntityIndex = resolveEntityToEntityIndex(ecs.Entities, e)
		ret = append(ret, entry)
	}
	flyingEntities := map[*ecs2.Entity]SessionEntity{}
	for _, e0 := range positionsAir.Results {
		for _, entry := range e0.Entries {
			if entry.Data == nil {
				continue
			}
			entity, ok := flyingEntities[entry.ResolvedEntity]
			if !ok {
				entity.PlayerID, entity.ModelName = resolveEntityDetails(players, entry.ResolvedEntity)
				entity.EntityIndex = resolveEntityToEntityIndex(ecs.Entities, entry.ResolvedEntity)
			}
			entity.Path = append(entity.Path, game.SpaceTime{
				Time: e0.CurrentTime,
				X:    float64(entry.Data.PosX),
				Y:    float64(entry.Data.PosY),
				Z:    float64(entry.Data.PosZ),
			})
			flyingEntities[entry.ResolvedEntity] = entity
		}
	}
	ret = append(ret, slices.Collect(maps.Values(flyingEntities))...)
	return
}
