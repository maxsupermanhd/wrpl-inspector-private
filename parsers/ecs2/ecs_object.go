package ecs2

type entity_id_t uint32

const ENTITY_INDEX_BITS = 22
const ENTITY_INDEX_MASK = (1 << ENTITY_INDEX_BITS) - 1
const ECS_INVALID_ENTITY_ID_VAL entity_id_t = 0

type EntityId struct { // TODO, maybe update other stuff to use this instead of uint64 TOMFOOLERY
	handle entity_id_t
}

func (id EntityId) Index() uint32 { // we probably only need index
	return uint32(id.handle & ENTITY_INDEX_MASK)
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
