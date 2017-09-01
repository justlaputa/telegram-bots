package main

import (
	"log"
	"os"
	"strings"
	"unicode/utf8"

	"cloud.google.com/go/translate"

	"golang.org/x/net/context"
	"golang.org/x/text/language"

	"google.golang.org/api/option"

	"gopkg.in/telegram-bot-api.v4"
)

const (
	MIN_MESSAGE_LENGTH    = 6
	MAX_MESSAGE_LENGTH    = 60
	UNKNOWN_USER_SPEAKING = "who is speaking?"
)

type Translation struct {
	ResultLanguage language.Tag
	TranslatedText string
}

type BotReply struct {
	Title        string
	Translations []Translation
}

func (reply BotReply) String() string {
	result := ""
	for _, translation := range reply.Translations {
		result = result + translation.ResultLanguage.String() + ": " + translation.TranslatedText + "\n"
	}
	return result
}

func detectLanguage(text, apiKey string) (language.Tag, error) {
	ctx := context.Background()
	client, err := translate.NewClient(ctx, option.WithAPIKey(apiKey))

	if err != nil {
		log.Printf("failed to create translate client object: %v", err)
		return language.Tag{}, err
	}

	resp, err := client.DetectLanguage(ctx, []string{text})
	if err != nil {
		log.Printf("failed to detect language from gcloud api: %v", err)
		return language.Tag{}, err
	}

	log.Printf("detecting language result: %+v", resp)

	return resp[0][0].Language, nil
}

func translateText(targetLanguage language.Tag, text, apiKey string) (Translation, error) {
	ctx := context.Background()

	client, err := translate.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		log.Printf("failed to create translate client object: %v", err)
		return Translation{}, err
	}

	resp, err := client.Translate(ctx, []string{text}, targetLanguage, nil)
	if err != nil {
		log.Printf("failed to translate by gcloud api: %v", err)
		return Translation{}, err
	}

	log.Printf("got all translations: %+v", resp)

	return Translation{targetLanguage, resp[0].Text}, nil
}

func getReplyTitle(fromUser string, sourceLanguage language.Tag) string {
	return ""
}

func getOtherLanguages(sourceLanguage language.Tag) []language.Tag {
	switch sourceLanguage {
	case language.Chinese:
		return []language.Tag{language.Japanese, language.English}
	case language.English:
		return []language.Tag{language.Japanese, language.Chinese}
	case language.Japanese:
		return []language.Tag{language.English, language.Chinese}
	default:
		return []language.Tag{language.Japanese, language.English, language.Chinese}
	}
}

func getReplyTranslations(fromUser string, sourceLanguage language.Tag, message, apiKey string) []Translation {
	targetLanguages := getOtherLanguages(sourceLanguage)
	translations := []Translation{}

	for _, lang := range targetLanguages {
		translation, err := translateText(lang, message, apiKey)
		if err != nil {
			log.Printf("failed to translate into %s", lang)
		} else {
			translations = append(translations, translation)
		}
	}

	return translations
}

func isGoodLength(message string) bool {
	len := utf8.RuneCountInString(message)
	return len > MIN_MESSAGE_LENGTH && len < MAX_MESSAGE_LENGTH
}

func isShort(message string) bool {
	len := utf8.RuneCountInString(message)
	return len <= MIN_MESSAGE_LENGTH
}

func containsFun(message string) bool {
	message = strings.ToLower(message)
	return strings.Contains(message, "cat") ||
		strings.Contains(message, "dog") ||
		strings.Contains(message, "ねこ") ||
		strings.Contains(message, "猫") ||
		strings.Contains(message, "犬") ||
		strings.Contains(message, "いぬ")
}

func isCommand(message string) bool {
	return strings.HasPrefix(message, "/")
}

func isUrl(message string) bool {
	return strings.HasPrefix(message, "http://") || strings.HasPrefix(message, "https://")
}

func processMessage(fromUser, message, apiKey string) (needReply bool, reply BotReply) {

	if !isGoodLength(message) || isCommand(message) || isUrl(message) {
		log.Printf("message is not worth processing, either too short or is a command or is url, I will skip it")
		return false, BotReply{}
	}

	sourceLanguage, err := detectLanguage(message, apiKey)
	if err != nil {
		return false, BotReply{}
	}

	log.Printf("detected message language: %s", sourceLanguage)

	reply = BotReply{}

	reply.Title = getReplyTitle(fromUser, sourceLanguage)
	reply.Translations = getReplyTranslations(fromUser, sourceLanguage, message, apiKey)

	return true, reply
}

func main() {
	token := os.Getenv("BOT_TOKEN")
	if token == "" {
		log.Fatal("can not find bot api token, did you set BOT_TOKEN in environment varialbe?")
	}

	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		log.Fatal("can not find gcloud translate api key, did you set API_KEY in environment varialbe?")
	}

	// bingImageSubKey := os.Getenv("BING_IMAGE_KEY")
	// if bingImageSubKey == "" {
	// 	log.Fatal("can not find bing image search subscription key, did you set BING_IMAGE_KEY in environment variable?")
	// }
	//
	// bingImageProvider := NewBingImageSearchProvider(bingImageSubKey)

	pinterestToken := os.Getenv("PINTEREST_TOKEN")
	if pinterestToken == "" {
		log.Fatal("can not find pinterest api token, did you set PINTEREST_TOKEN in environment varialbe?")
	}

	pinterest := &Pinterest{APIToken: pinterestToken}

	log.Println("starting translate bot with specified token...")

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

		if (isShort(update.Message.Text) && containsFun(update.Message.Text)) ||
			strings.HasPrefix(update.Message.Text, "p ") || strings.HasPrefix(update.Message.Text, "P ") {
			query := update.Message.Text[2:]
			imageURL, err := pinterest.Search(query)
			if err != nil {
				log.Printf("failed to get image, skip silently")
				continue
			}

			replyImage := tgbotapi.PhotoConfig{
				BaseFile: tgbotapi.BaseFile{
					BaseChat: tgbotapi.BaseChat{
						ChatID:           update.Message.Chat.ID,
						ReplyToMessageID: update.Message.MessageID,
					},
					FileID:      imageURL,
					UseExisting: true,
					MimeType:    "image/jpeg",
				},
			}
			message, err := bot.Send(replyImage)
			if err != nil {
				log.Printf("failed to send reply, %v", err)
			} else {
				log.Printf("replied successful: %+v", message)
			}
		} else {

			needReply, reply := processMessage(update.Message.From.LastName, strings.TrimSpace(update.Message.Text), apiKey)

			if needReply {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, reply.String())
				msg.ReplyToMessageID = update.Message.MessageID

				bot.Send(msg)
			}
		}
	}
}
