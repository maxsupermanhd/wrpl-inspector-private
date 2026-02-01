package ecs2

import "github.com/maxsupermanhd/wrpl-inspector/v2/wrpl/danet"

type ComponentParser interface {
	Parse(r *danet.BitReader, ctx *PacketECSParser) (ret any, err error)
}
