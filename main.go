package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/halink0803/telegram-unsplash-bot/common"
	"github.com/halink0803/telegram-unsplash-bot/unsplash"
)

func readConfigFromFile(path string) (common.BotConfig, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return common.BotConfig{}, err
	}
	result := common.BotConfig{}
	err = json.Unmarshal(data, &result)
	return result, err
}

//Bot the main bot
type Bot struct {
	bot      *tgbotapi.BotAPI
	unsplash *unsplash.Unsplash
}

func main() {
	configPath := "config.json"
	botConfig, err := readConfigFromFile(configPath)
	if err != nil {
		log.Panic(err)
	}

	// init bot
	bot, err := tgbotapi.NewBotAPI(botConfig.BotKey)
	if err != nil {
		log.Panic(err)
	}
	bot.Debug = true

	// init unsplash client
	unsplash := unsplash.NewUnsplash(botConfig.UnsplashKey, botConfig.UnsplashSecret)

	mybot := Bot{
		bot:      bot,
		unsplash: unsplash,
	}

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := mybot.bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		log.Printf("Updates:[%s] %s", update.Message.From.UserName, update.Message.Command())

		mybot.handle(update)

		// msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
		// msg.ReplyToMessageID = update.Message.MessageID

		// mybot.bot.Send(msg)
	}
}

func (mybot *Bot) handle(update tgbotapi.Update) {
	if update.Message.IsCommand() {
		switch update.Message.Command() {
		case "search":
			mybot.handleSearch(update)
			break
		default:
			break
		}
	}
}

func (mybot *Bot) handleSearch(update tgbotapi.Update) {
	arguments := update.Message.CommandArguments()
	if len(arguments) == 0 {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "What do you want to search?")
		mybot.bot.Send(msg)
		return
	}
	results, err := mybot.unsplash.SearchPhotos(arguments)
	if err != nil {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Cannot search photos: %s", err.Error()))
		mybot.bot.Send(msg)
	}
	for _, photo := range results.Results {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, photo.URLs.Regular)
		mybot.bot.Send(msg)
	}
}
