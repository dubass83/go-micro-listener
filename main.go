package main

import (
	"fmt"
	"math"
	"os"
	"time"

	"github.com/dubass83/go-micro-listener/util"
	amqp "github.com/rabbitmq/amqp091-go"
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

	rabbitConn, err := tryConnectToRabbitIn(5, conf)
	if err != nil {
		log.Fatal().Msg("failed to establish the connection to rabbitmq")
	}
	defer rabbitConn.Close()

}

func tryConnectToRabbitIn(backOffLimit int64, conf util.Config) (*amqp.Connection, error) {
	var counts int64
	backOff := 1 * time.Second
	var connection *amqp.Connection

	// wait until rabbit is ready
	for {
		// check if we reach the backOffLimit the stop trying and return an error
		if counts > backOffLimit {
			msg := fmt.Sprintf("failed connect to rabbit with %d tries", backOffLimit)
			log.Error().Msg(msg)
			return nil, fmt.Errorf(msg)
		}
		// try connect to rabbitmq
		c, err := amqp.Dial(conf.RabbitURL)
		if err != nil {
			log.Error().Err(err).Msg("rabbitmq not yet ready...")
			time.Sleep(backOff)
			counts++
			backOff = time.Duration(math.Pow(float64(counts), 2)) * time.Second
			continue
		}
		log.Debug().Msg("connected to rabbitmq")
		connection = c
		break
	}
	return connection, nil
}
