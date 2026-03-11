package carve

import (
	"slices"
	"wrplinspectorprivate/parsers/critical"
	"wrplinspectorprivate/parsers/ecs2"
	"wrplinspectorprivate/parsers/severe"
	"wrplinspectorprivate/parsers/slot2"
)

type DamageVariant byte

const (
	DamageVariantHit DamageVariant = iota
	DamageVariantCritical
	DamageVariantSevere
)

type SessionDamage struct {
	Time                uint32
	Variant             DamageVariant
	OffenderID          uint64
	OffenderModel       string
	OffenderEntityIndex uint32
	OffendedID          uint64
	OffendedModel       string
	OffendedEntityIndex uint32
	CausedFire          bool `json:",omitempty"`
}

func assembleDamage(
	ecs *ecs2.EntityManager,
	players [256]*slot2.Player,
	damageCritical *critical.CriticalDamageParser,
	damageSevere *severe.SevereDamageParser,
) (ret []SessionDamage, err error) {
	for _, d := range damageCritical.Results {
		entry := SessionDamage{
			Time:       d.CurrentTime,
			Variant:    DamageVariantCritical,
			CausedFire: d.Fire,
		}
		if d.OffendedEntity != nil {
			entry.OffendedID, entry.OffendedModel = resolveEntityDetails(players, d.OffendedEntity)
			entry.OffendedEntityIndex = resolveEntityToEntityIndex(ecs.Entities, d.OffendedEntity)
		}
		if d.PlayerEntity != nil {
			entry.OffenderID, entry.OffenderModel = resolveEntityDetails(players, d.PlayerEntity)
			entry.OffenderEntityIndex = resolveEntityToEntityIndex(ecs.Entities, d.PlayerEntity)
		} else {
			entry.OffenderModel = d.Vehicle
			if d.PlayerPID >= uint32(len(players)) {
				continue
			}
			player := players[d.PlayerPID]
			if player == nil {
				continue
			}
			entry.OffenderID = uint64(player.UserID)
		}
		ret = append(ret, entry)
	}
	for _, d := range damageSevere.Results {
		entry := SessionDamage{
			Time:    d.CurrentTime,
			Variant: DamageVariantSevere,
		}
		if d.OffendedEntity != nil {
			entry.OffendedID, entry.OffendedModel = resolveEntityDetails(players, d.OffendedEntity)
			entry.OffendedEntityIndex = resolveEntityToEntityIndex(ecs.Entities, d.OffendedEntity)
		}
		if d.PlayerEntity != nil {
			entry.OffenderID, entry.OffenderModel = resolveEntityDetails(players, d.PlayerEntity)
			entry.OffenderEntityIndex = resolveEntityToEntityIndex(ecs.Entities, d.PlayerEntity)
		}
		ret = append(ret, entry)
	}
	slices.SortFunc(ret, func(a, b SessionDamage) int {
		return int(a.Time) - int(b.Time)
	})
	return
}
