package main

import (
	"log"
	"net/http"
	"os"
	"strings"

	"gopkg.in/telegram-bot-api.v4"
)

func main() {
	token := os.Getenv("BOT_TOKEN")
	if token == "" {
		log.Fatalln("BOT_TOKEN is not set, try to set it in environment variable and start again")
	}

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatalf("failed to create telegram bot, %v", err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	_, err = bot.SetWebhook(tgbotapi.NewWebhook("https://apps.laputa.io/cat"))
	if err != nil {
		log.Fatalf("something wrong while set webhook to telegram bot, %v", err)
	}

	updates := bot.ListenForWebhook("/cat")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8086"
	}

	log.Printf("starting server on :%s...\n", port)

	go http.ListenAndServe("127.0.0.1:"+port, nil)

	for update := range updates {
		log.Printf("%+v\n", update)
		log.Printf("got message: %+v", update.Message)

		if strings.HasPrefix(update.Message.Text, "/cat") || strings.HasPrefix(update.Message.Text, "/dog") {
			log.Printf("got %s request from %s", update.Message.Text, update.Message.From.UserName)

			chatID := update.Message.Chat.ID
			log.Printf("cat id: %d", chatID)

			replyAction := tgbotapi.ChatActionConfig{
				BaseChat: tgbotapi.BaseChat{
					ChatID: chatID,
				},
				Action: "upload_photo",
			}

			log.Printf("sending photo upload action: %v", replyAction)
			message, err := bot.Send(replyAction)
			if err != nil {
				log.Printf("failed to send chat action: %v, %v", message, err)
			}
			log.Printf("returned message: %v+", message)

			var subject string
			if strings.HasPrefix(update.Message.Text, "/cat") {
				subject = "cats"
			} else {
				subject = "cute"
			}

			pinterestResponse, err := PintrestInterests(subject)
			if err != nil {
				log.Printf("failed to get cat image from pinterest: %v", err)
				continue
			}

			replyImage := tgbotapi.PhotoConfig{
				BaseFile: tgbotapi.BaseFile{
					BaseChat: tgbotapi.BaseChat{
						ChatID:           chatID,
						ReplyToMessageID: update.Message.MessageID,
					},
					FileID:      pinterestResponse.Data.Images.Orig.URL,
					UseExisting: true,
					MimeType:    "image/jpeg",
				},
				Caption: pinterestResponse.Data.Attribution.Title,
			}
			message, err = bot.Send(replyImage)
			if err != nil {
				log.Printf("failed to send reply, %v", err)
			} else {
				log.Printf("replied successful: %+v", message)
			}
		}
	}
}
