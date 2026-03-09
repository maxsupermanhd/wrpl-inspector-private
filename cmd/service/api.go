package main

import (
	"archive/tar"
	"bytes"
	"encoding/gob"
	"encoding/json"
	"io"
	"main/carve"
	"net/http"
	"os"

	"github.com/rs/zerolog"
)

func apiCarveBundle(w http.ResponseWriter, r *http.Request, logvals *zerolog.Event) {
	ret, err := carve.CarveBundle(tar.NewReader(r.Body), carve.CarveParams{}, *parserECSHashes)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write([]byte(err.Error()))
		return
	}
	apiReplyWithTar(w, ret)
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
	apiReplyWithTar(w, ret)
}

func apiReplyWithTar(w http.ResponseWriter, ret *carve.CarvedReplay) {
	carveJSON, err := json.Marshal(ret)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	carveGOB := &bytes.Buffer{}
	err = gob.NewEncoder(carveGOB).Encode(ret)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	w.WriteHeader(http.StatusOK)
	tarWriter := tar.NewWriter(w)
	tarWriter.WriteHeader(&tar.Header{
		Typeflag: tar.TypeReg,
		Name:     "carve.json",
		Size:     int64(len(carveJSON)),
	})
	tarWriter.Write(carveJSON)
	tarWriter.WriteHeader(&tar.Header{
		Typeflag: tar.TypeReg,
		Name:     "carve.gob",
		Size:     int64(carveGOB.Len()),
	})
	tarWriter.Write(carveGOB.Bytes())
	tarWriter.Close()
}
