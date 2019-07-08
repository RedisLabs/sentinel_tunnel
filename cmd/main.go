package main

import (
	"fmt"
	"github.com/DivPro/sentinel_tunnel/cmd/config"
	"github.com/DivPro/sentinel_tunnel/cmd/resolver"
	"github.com/DivPro/sentinel_tunnel/cmd/server"
	"github.com/rs/zerolog/log"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage : sentinel_tunnel <config_file_path>")
		return
	}

	conf, err := config.CreateFromFile(os.Args[1])
	if err != nil {
		log.Fatal().Err(err).Msg("init config")
	}

	sentinels, err := resolver.CreateSentinels(conf.Sentinels)
	if err != nil {
		log.Fatal().Err(err).Msg("connect to sentinels")
	}
	srv := server.NewServer(
		resolver.NewResolver(sentinels),
		conf.Databases,
	)
	go func() {
		err := srv.Start()
		if err != nil {
			log.Fatal().Err(err).Msg("starting server")
		}
	}()

	ctx := server.GetShutdownCtx()
	<-ctx.Done()
	srv.Stop()
}
