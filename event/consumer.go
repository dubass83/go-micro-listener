package event

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog/log"
)

const (
	LogService = "http://logger:8080"
)

type Consumer struct {
	conn      *amqp.Connection
	queueName string
}

type Payload struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

type jsonResponse struct {
	Error   bool   `json:"error"`
	Massage string `json:"massage"`
	Data    any    `json:"data,omitempty"`
}

func NewConsumer(conn *amqp.Connection) (Consumer, error) {
	consumer := Consumer{
		conn: conn,
	}
	if err := consumer.setup(); err != nil {
		log.Error().Err(err).Msg("failed to setup consumer")
		return Consumer{}, err
	}

	return consumer, nil
}

func (consumer *Consumer) setup() error {
	channel, err := consumer.conn.Channel()
	if err != nil {
		log.Error().Err(err).Msg("failed to create chanel for consumer")
		return err
	}
	return declareExchange(channel)
}

func (consumer *Consumer) Listen(topics []string) error {
	ch, err := consumer.conn.Channel()
	if err != nil {
		log.Error().Err(err).Msg("failed to get chanel for consumer")
		return err
	}
	defer ch.Close()

	q, err := declareRandomQueue(ch)
	if err != nil {
		log.Error().Err(err).Msg("failed to create random queue for consumer")
		return err
	}

	for _, s := range topics {
		err := ch.QueueBind(
			q.Name,
			s,
			"logs_topic",
			false,
			nil,
		)

		if err != nil {
			log.Error().Err(err).Msg("failed to bind an exchange to the queue")
			return err
		}
	}

	messages, err := ch.Consume(q.Name, "", true, false, false, false, nil)
	if err != nil {
		log.Error().Err(err).Msg("failed to consume messages from the queue")
		return err
	}

	forever := make(chan bool)
	go func() {
		for d := range messages {
			var payload Payload
			_ = json.Unmarshal(d.Body, &payload)
			go handlePayload(payload)
		}
	}()
	log.Info().Msgf("waiting for message [Exchange, Queue] [logs_topics, %s]", q.Name)
	<-forever

	return nil
}

func handlePayload(payload Payload) {
	switch payload.Name {

	case "log", "event":
		log.Info().Msg("send event to the logger service")
		err := logEvent(payload)
		if err != nil {
			log.Error().Err(err).Msg("failed to send event to the logger service")
		}

	case "auth":
		log.Info().Msg("send event to the auth service")

	default:
		log.Info().Msg("send event to the logger service as default case")
		err := logEvent(payload)
		if err != nil {
			log.Error().Err(err).Msg("failed to send event to the logger service")
		}
	}
}

func logEvent(logs Payload) error {
	log.Debug().Msg("post log into logger service")
	jsonData, _ := json.MarshalIndent(logs, "", "\t")
	logURL := fmt.Sprintf("%s/log", LogService)
	log.Debug().Msgf("logURL: %s", logURL)
	request, err := http.NewRequest("POST", logURL, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Error().Err(err).Msg("failed generate POST request")
		return err
	}

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		log.Error().Err(err).Msg("failed to create http Client")
		return err
	}
	defer response.Body.Close()

	// if response.StatusCode == http.StatusUnauthorized {
	// 	errorJSON(w, errors.New("invalid credentials"))
	// 	return
	// }
	if response.StatusCode != http.StatusAccepted {
		log.Error().Err(err).Msg("return bad status code in response")
		return err
	}

	return nil
}
