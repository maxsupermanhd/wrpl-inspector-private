package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"math/rand/v2"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func makeHTTPServeMux() http.HandlerFunc {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", httpLog(handle404))
	mux.HandleFunc("GET /testpage", serveTestpage)

	// mux.HandleFunc("GET /{$}", httpLog(serveIndex))

	mux.HandleFunc("POST /api/0/carve/bundle", httpLogWithEvent(apiCarveBundle))
	mux.HandleFunc("POST /api/0/carve/path", httpLogWithEvent(apiCarvePath))
	// mux.HandleFunc("GET /api/0/ws", websocket.Server{Handler: wsAPI}.ServeHTTP)

	return mux.ServeHTTP
}

func httpLogGetRemoteIP(r *http.Request) string {
	ip := r.Header.Get("CF-Connecting-IP")
	if ip == "" {
		li := strings.LastIndex(r.RemoteAddr, ":")
		if li == -1 {
			ip = r.RemoteAddr
		} else {
			ip = r.RemoteAddr[:li]
		}
	}
	return ip
}

func httpLog(f func(w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()
		w2 := &httpResponseWriterCapturer{ResponseWriter: w}
		ip := httpLogGetRemoteIP(r)
		// auth := cfg.GetDString("", append([]string{"auth"}, r.Header.Get("Authorization"), "name")...)
		auth := ""
		log.Info().Msgf("%2s %15s %3d %7s %20q %q %q %q",
			r.Header.Get("CF-IPCountry"), ip,
			w2.lastStatus, "hit", r.Header.Get("Cf-Ray"),
			r.URL.Path, r.UserAgent(), auth)
		f(w2, r)
		log.Info().Msgf("%2s %15s %3d %7s %20q %q %q %q",
			r.Header.Get("CF-IPCountry"), ip,
			w2.lastStatus, time.Since(startTime).Round(1*time.Millisecond).String(), r.Header.Get("Cf-Ray"),
			r.URL.Path, r.UserAgent(), auth)
	}
}

func httpLogWithEvent(f func(w http.ResponseWriter, r *http.Request, logvals *zerolog.Event)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()
		w2 := &httpResponseWriterCapturer{ResponseWriter: w}
		ip := httpLogGetRemoteIP(r)
		// auth := cfg.GetDString("", append([]string{"auth"}, r.Header.Get("Authorization"), "name")...)
		auth := ""
		log.Info().Msgf("%2s %15s %3d %7s %20q %q %q %q",
			r.Header.Get("CF-IPCountry"), ip,
			w2.lastStatus, "hit", r.Header.Get("Cf-Ray"),
			r.URL.Path, r.UserAgent(), auth)
		le := log.Info()
		f(w2, r, le)
		le.Msgf("%2s %15s %3d %7s %20q %q %q %q",
			r.Header.Get("CF-IPCountry"), ip,
			w2.lastStatus, time.Since(startTime).Round(1*time.Millisecond).String(), r.Header.Get("Cf-Ray"),
			r.URL.Path, r.UserAgent(), auth)
	}
}

func httpRoutine(exitChan <-chan struct{}) {
	listenAddr := "127.0.0.1:63332"
	serv := http.Server{
		Addr:              listenAddr,
		Handler:           makeHTTPServeMux(),
		ReadTimeout:       time.Minute,
		WriteTimeout:      time.Minute,
		IdleTimeout:       time.Minute,
		ReadHeaderTimeout: 3 * time.Second,
	}

	wg := &sync.WaitGroup{}
	wg.Go(func() {
		log.Info().Str("addr", listenAddr).Msg("http listening")
		err := serv.ListenAndServe()
		if !errors.Is(err, http.ErrServerClosed) {
			log.Err(err).Msg("server closed")
		}
	})

	<-exitChan

	ctx, ctxc := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxc()
	serv.Shutdown(ctx)
}

func handle404(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	fmt.Fprintf(w, "path %q has nothing to offer\n\n", r.URL.Path)
}

type httpResponseWriterCapturer struct {
	http.ResponseWriter
	lastStatus int
	http.Hijacker
}

func (rc *httpResponseWriterCapturer) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return rc.ResponseWriter.(http.Hijacker).Hijack()
}

func (rc *httpResponseWriterCapturer) Unwrap() http.ResponseWriter {
	return rc.ResponseWriter
}

func (rc *httpResponseWriterCapturer) WriteHeader(statusCode int) {
	rc.lastStatus = statusCode
	rc.ResponseWriter.WriteHeader(statusCode)
}

func serveTestpage(w http.ResponseWriter, _ *http.Request) {
	rng := rand.NewPCG(uint64(time.Now().Unix()), uint64(time.Now().Unix())-69420)
	ret := &strings.Builder{}
	chars := "0123456789qwertyuiopasdfghjklzxcvbnm"
	for range 200_000 {
		ret.WriteByte(chars[int(rng.Uint64()%6251)%len(chars)])
	}
	w.WriteHeader(200)
	w.Write([]byte(ret.String()))
}
