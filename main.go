package main

import (
	"log"
	"fmt"
	"os"

	"encoding/json"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/nats-io/nats.go"
)

func main() {
	// Init

	var appName, tgToken, natsAddr, natsPubPrefix string
	var err error
	var ok bool

	appName, ok = os.LookupEnv("APP_NAME")
	if !ok {
		log.Panic("set application name")
	}

	tgToken, ok = os.LookupEnv("APP_TG_TOKEN")
	if !ok {
		log.Panic("can't start without Telegram token")
	}

	natsAddr, ok = os.LookupEnv("APP_NATS_ADDR")
	if !ok {
		natsAddr = nats.DefaultURL
	}

	natsPrefix, ok = os.LookupEnv("APP_NATS_PREFIX")
	if !ok {
		natsPrefix = "dummy"
	}

	// NATS

	nc, err := nats.Connect(natsAddr)
	if err != nil {
		log.Panic("could not connect: ", err)
	}

	defer nc.Close()
	defer nc.Flush()

	js, err := nc.JetStream()
	if err != nil {
		log.Panic("could not get jetstream: ", err)
	}

	natsStreamChannel := fmt.Sprintf("%s.tg.in", natsPrefix)

	js.AddStream(&nats.StreamConfig{
		Name:     appName,
		Subjects: []string{
			natsStreamChannel,
			fmt.Sprintf("%s.*", natsStreamChannel),
		},
		Discard:  nats.DiscardOld,
		MaxMsgs:  1000,
	})

	js.AddConsumer("worker", &nats.ConsumerConfig{})

	// Telegram

	bot, err := tgbotapi.NewBotAPI(tgToken)
	if err != nil {
		log.Panic("could not get telegram: ", err)
	}

	// Run logic

	// bot.Debug = true

	log.Printf("authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil { // If we got a message
			log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

			messageData, err := json.Marshal(update.Message)
			chatID := update.Message.Chat.ID

			// TODO: Backoff until success
			if err != nil {
				log.Println(err)
				continue
			}

			natsPubChannel := fmt.Sprintf("%s.%d", natsStreamChannel, chatID)

			log.Println("publish to", natsPubChannel)

			_, err = js.Publish(natsPubChannel, messageData)
			if err != nil {
				log.Println(err)
				continue
			}
		}
	}
}
