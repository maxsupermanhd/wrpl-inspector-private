package main

import (
	"bytes"
	"context"
	"flag"
	"os"
	"os/signal"
	"sync"
	"wrplinspectorprivate/parsers/ecs2"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	flECSHashesJSONPath = flag.String("ecshashes", "../../ecshashes.json", "path to ecshashes.json file")
	parserECSHashes     *ecs2.ComponentHashMaps
)

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	flag.Parse()
	parserECSHashes = noerr(ecs2.ReadComponentHashMaps(bytes.NewReader(noerr(os.ReadFile(*flECSHashesJSONPath)))))

	ctx, ctxClose := signal.NotifyContext(context.Background(), os.Interrupt)
	defer ctxClose()

	wg := sync.WaitGroup{}
	wg.Go(func() {
		httpRoutine(ctx.Done())
	})

	<-ctx.Done()
	log.Info().Msg("Got sigterm, shutting down")

	wg.Wait()

	log.Info().Msg("bye")
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
