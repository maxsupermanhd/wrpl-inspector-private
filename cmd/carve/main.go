package main

import (
	"archive/tar"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"main/carve"
	"main/parsers/ecs2"
	"os"
	"strconv"
)

var (
	flECSHashesJSONPath = flag.String("ecshashes", "../../ecshashes.json", "path to ecshashes.json file")
	flSample            = flag.Int("sample", 0, "trim all arrays to n elements")
	parserECSHashes     *ecs2.ComponentHashMaps
)

func main() {
	flag.Parse()
	parserECSHashes = noerr(ecs2.ReadComponentHashMaps(bytes.NewReader(noerr(os.ReadFile(*flECSHashesJSONPath)))))

	if len(flag.Args()) == 0 {
		fmt.Println("no inputs")
		return
	}

	for _, arg := range flag.Args() {
		f := noerr(os.ReadFile(arg))
		carved := noerr(carve.CarveBundle(tar.NewReader(bytes.NewReader(f)), carve.CarveParams{}, *parserECSHashes))
		if *flSample > 0 {
			carved.Players = carved.Players[:min(*flSample, len(carved.Players))]
			carved.Kills = carved.Kills[:min(*flSample, len(carved.Kills))]
			carved.Awards = carved.Awards[:min(*flSample, len(carved.Awards))]
			carved.DamageReports = carved.DamageReports[:min(*flSample, len(carved.DamageReports))]
			carved.Entities = carved.Entities[:min(*flSample, len(carved.Entities))]
		}
		carvedJSON := noerr(json.MarshalIndent(carved, "", "\t"))
		must(os.WriteFile(strconv.FormatUint(carved.SessionID, 10)+".json", carvedJSON, 0644))
	}
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
