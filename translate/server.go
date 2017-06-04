package main

import (
	"log"
	"os"

	"cloud.google.com/go/translate"

	"golang.org/x/net/context"
	"golang.org/x/text/language"
	"gopkg.in/telegram-bot-api.v4"

	"google.golang.org/api/option"
)

const (
	MINIMUM_MESSAGE_LENGTH = 3
	DOG_NAME               = "justlaputa"
	CAT_NAME               = "Masami"
	DOG_USE_ENGLISH        = "bad dog, why use english? let smart dog teach you"
	DOG_USE_CHINESE        = "umm, seems dog is teaching chinese, smart dog also know Chinese"
	DOG_USE_JAPANESE       = "good dog, you are using Japanese! let smart dog check it"
	CAT_USE_ENGLISH        = "bad cat, why you use English? you need teach dog Japanese! let smart dog help you"
	CAT_USE_JAPANESE       = ""
	CAT_USE_CHINESE        = "wow, cat is talking Chinese, let smart dog check"
	UNKNOWN_USER_SPEAKING  = "who is speaking?"
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
	result := reply.Title + "\n"
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

	return Translation{targetLanguage, resp[0].Text}, nil
}

func getReplyTitle(fromUser string, sourceLanguage language.Tag) string {
	var reply string
	if fromUser == DOG_NAME {
		switch sourceLanguage {
		case language.English:
			reply = DOG_USE_ENGLISH
		case language.Chinese:
			reply = DOG_USE_CHINESE
		case language.Japanese:
			reply = DOG_USE_JAPANESE
		default:
			reply = ""
		}
	} else if fromUser == CAT_NAME {
		switch sourceLanguage {
		case language.English:
			reply = CAT_USE_ENGLISH
		case language.Chinese:
			reply = CAT_USE_CHINESE
		case language.Japanese:
			reply = CAT_USE_JAPANESE
		default:
			reply = ""
		}
	} else {
		reply = UNKNOWN_USER_SPEAKING
	}
	return reply
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

func processMessage(fromUser, message, apiKey string) (needReply bool, reply BotReply) {
	if len(message) <= MINIMUM_MESSAGE_LENGTH {
		log.Printf("message is too short, probably no need to process, I will skip it")
		return false, BotReply{}
	}

	sourceLanguage, err := detectLanguage(message, apiKey)
	if err != nil {
		return false, BotReply{}
	}

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

		needReply, reply := processMessage(update.Message.From.UserName, update.Message.Text, apiKey)

		if needReply {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, reply.String())
			msg.ReplyToMessageID = update.Message.MessageID

			bot.Send(msg)
		}
	}
}