package ecs2

import "github.com/maxsupermanhd/wrpl-inspector/wrpl/danet"

type ComponentParser interface {
	Parse(r *danet.BitReader) (ret any, err error)
}
