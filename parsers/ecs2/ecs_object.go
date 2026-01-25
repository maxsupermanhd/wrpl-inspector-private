package ecs2

const ENTITY_INDEX_BITS = 22
const ENTITY_INDEX_MASK = (1 << ENTITY_INDEX_BITS) - 1
const ECS_INVALID_ENTITY_ID_VAL EntityID = 0

type EntityID uint32

func (id EntityID) Index() uint32 { // we probably only need index
	return uint32(id & ENTITY_INDEX_MASK)
}

type Component struct {
	Value any           // actual data, I think this will ensure the Component owns?
	Type  ComponentHash // what type
}

type Object struct {
	Components map[string]Component // maps a name to a specific component
}

type Entity struct {
	Template string
	Data     Object
}
