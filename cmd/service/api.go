package main

import (
	"archive/tar"
	"bytes"
	"encoding/hex"
	"encoding/json"
	"io"
	"main/carve"
	"net/http"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func apiCarveBundle(w http.ResponseWriter, r *http.Request, logvals *zerolog.Event) {
	f, _ := io.ReadAll(r.Body)
	if len(f) > 256 {
		log.Info().Msg("\n" + hex.Dump(f[:256]))
	} else {
		log.Info().Msg("\n" + hex.Dump(f))
	}
	ret, err := carve.CarveBundle(tar.NewReader(bytes.NewReader(f)), carve.CarveParams{}, *parserECSHashes)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write([]byte(err.Error()))
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(ret)
}

func apiCarvePath(w http.ResponseWriter, r *http.Request, logvals *zerolog.Event) {
	p, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	f, err := os.Open(string(p))
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write([]byte(err.Error()))
		return
	}
	defer f.Close()
	ret, err := carve.CarveBundle(tar.NewReader(f), carve.CarveParams{}, *parserECSHashes)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write([]byte(err.Error()))
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(ret)
}
