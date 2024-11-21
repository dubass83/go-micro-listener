package main

import (
	"os"

	"github.com/dubass83/go-micro-listener/util"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	conf, err := util.LoadConfig(".")
	if err != nil {
		log.Fatal().
			Err(err).
			Msg("cannot load configuration")
	}
	if conf.Enviroment == "devel" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
		// log.Debug().Msgf("config values: %+v", conf)
	}

}
