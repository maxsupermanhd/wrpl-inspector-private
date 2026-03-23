package main

import (
	"encoding/hex"
	"fmt"

	"github.com/maxsupermanhd/wrpl-inspector-private/idfieldserializer"
	"github.com/maxsupermanhd/wrpl-inspector/v2/wrpl/danet"
)

func main() {
	packetPayload := noerr(hex.DecodeString(`7068000460010001000100010001000100010001000100010001000100010001000100010001000100010001000201010000000000000000000000000000000000000000000000000000000000000000000000000000000000000000ffff000000000000000039a7271800498b0030`))
	r := danet.NewBitReader(packetPayload[1:])
	must(idfieldserializer.DeserializeIdFieldSerializer255(r, func(fieldIndex uint16, fieldSize uint32) error {
		fmt.Printf("%v %v\n%s\n", fieldIndex, fieldSize, hex.Dump(noerr(r.ReadBits(int(fieldSize)))))
		return nil
	}))
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
