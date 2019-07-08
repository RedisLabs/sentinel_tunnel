package server

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/rs/zerolog/log"
)

var (
	ctx, cancelFunc = context.WithCancel(context.Background())
	gc              sync.Once
	catcher         chan os.Signal
)

func init() {
	catcher = make(chan os.Signal, 1)
	signal.Notify(catcher, syscall.SIGINT, syscall.SIGTERM)
}

func GetShutdownCtx() context.Context {
	gc.Do(func() {
		log.Info().Msg("launch catch gorutine")
		go launchCatcher()
	})

	return ctx
}

func Shutdown() {
	cancelFunc()
}

func launchCatcher() {
	go func() {
		for sig := range catcher {
			switch sig {
			case syscall.SIGTERM:
				log.Info().Msg("Got SIGTERM stopping application")
				cancelFunc()
			case syscall.SIGINT:
				log.Info().Msg("Got SIGINT stopping application")
				cancelFunc()
			}

			log.Info().Msg("catch signal")
		}
	}()
}
