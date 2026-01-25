package ecs2

import (
	"encoding/json"
	"fmt"
	"io"
)

type DataComponentDef struct {
	Name          string `json:"name"`
	ComponentHash uint32 `json:"comp"`
	CustomLoader  bool   `json:"has_loader"`
}

type ComponentHashMaps struct {
	ComponentNames map[uint32]string           `json:"components"`
	DataComponents map[uint32]DataComponentDef `json:"dataComponents"`
}

func ReadComponentHashMaps(r io.Reader) (ret *ComponentHashMaps, err error) {
	type DataComponentDefStringed struct {
		Name          string `json:"name"`
		ComponentHash string `json:"comp"`
		CustomLoader  bool   `json:"has_loader"`
	}
	type ComponentHashMapsStringed struct {
		ComponentNames map[string]string                   `json:"components"`
		DataComponents map[string]DataComponentDefStringed `json:"dataComponents"`
	}
	var m ComponentHashMapsStringed
	err = json.NewDecoder(r).Decode(&m)
	if err != nil {
		return nil, err
	}
	ret = &ComponentHashMaps{
		ComponentNames: map[uint32]string{},
		DataComponents: map[uint32]DataComponentDef{},
	}
	for k, v := range m.ComponentNames {
		knum, err := parseHashStr(k)
		if err != nil {
			return nil, err
		}
		ret.ComponentNames[knum] = v
	}
	for k, v := range m.DataComponents {
		knum, err := parseHashStr(k)
		if err != nil {
			return nil, err
		}
		chnum, err := parseHashStr(v.ComponentHash)
		if err != nil {
			return nil, err
		}
		ret.DataComponents[knum] = DataComponentDef{
			Name:          v.Name,
			ComponentHash: chnum,
			CustomLoader:  v.CustomLoader,
		}
	}
	return
}

func parseHashStr(k string) (ret uint32, err error) {
	var n int
	n, err = fmt.Sscanf(k, "0x%x", &ret)
	if n != 1 {
		return ret, fmt.Errorf("not scanned hex number: %q", k)
	}
	return ret, nil
}
