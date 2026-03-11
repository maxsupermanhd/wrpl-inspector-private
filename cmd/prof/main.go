package main

import (
	"bytes"
	"fmt"
	"os"
	"runtime/pprof"
	"time"

	"github.com/maxsupermanhd/wrpl-inspector-private/carve"
	"github.com/maxsupermanhd/wrpl-inspector-private/parsers/ecs2"

	"github.com/maxsupermanhd/wrpl-inspector/v2/wrpl"
)

func main() {

	cpuprofBytes := &bytes.Buffer{}
	must(pprof.StartCPUProfile(cpuprofBytes))

	p := "/home/max/.var/app/com.valvesoftware.Steam/.local/share/Steam/steamapps/common/War Thunder/Replays/#2026.03.07 00.15.18.wrpl"
	chms := noerr(ecs2.ReadComponentHashMaps(bytes.NewReader(noerr(os.ReadFile("../../ecshashes.json")))))

	replayBytes := noerr(os.ReadFile(p))

	perfSum := time.Duration(0)
	perfCount := 200

	for i := range perfCount {
		perfTime := time.Now()
		rr := noerr(wrpl.OpenReplay(bytes.NewReader(replayBytes), true, true, true))
		defer rr.Close()
		noerr(carve.CarveReplay(map[int]*wrpl.ReplayReader{0: rr}, *chms))
		perfDur := time.Since(perfTime)
		perfSum += perfDur
		fmt.Println(perfDur, perfSum/time.Duration(i+1))
	}

	pprof.StopCPUProfile()
	must(os.WriteFile("cpu.prof", cpuprofBytes.Bytes(), 0644))

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
