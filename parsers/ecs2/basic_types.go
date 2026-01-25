package ecs2

type ComponentHash uint32
type DataComponentHash uint32

type Point2 struct {
	x float32
	y float32
}

type Point3 struct {
	x float32
	y float32
	z float32
}
type Point4 struct {
	x float32
	y float32
	z float32
	w float32
}

type IPoint2 struct {
	x int32
	y int32
}

type IPoint3 struct {
	x int32
	y int32
	z int32
}

type IPoint4 struct {
	x int32
	y int32
	z int32
	w int32
}

type TMatrix struct {
	v1 Point3
	v2 Point3
	v3 Point3
	v4 Point3
}

type E3DColor struct {
	a uint8
	r uint8
	g uint8
	b uint8
}

type RendInstDesc struct {
	v1 uint32 // actually 32
	v2 uint32 // actually 32
	v3 uint32 // actually 32
}

type PartId uint8 // game only reads 6 bits

type riex_handle_t uint64

type UnitStorage struct {
	raw []byte
}

type Rocket struct { // dummy
	womp uint8
}
