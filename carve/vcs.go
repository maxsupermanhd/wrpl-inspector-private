package carve

import (
	"runtime/debug"
	"strings"
)

var (
	vcsReport map[string]string
)

func init() {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return
	}
	for _, v := range info.Settings {
		if strings.HasPrefix(v.Key, "vcs") {
			vcsReport[v.Key] = v.Value
		}
	}
}
