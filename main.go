package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/halink0803/telegram-unsplash-bot/common"
	"github.com/halink0803/telegram-unsplash-bot/unsplash"
	"google.golang.org/appengine"
	"google.golang.org/appengine/urlfetch"
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
	log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.Lshortfile)

	configPath := "config.json"
	botConfig, err := readConfigFromFile(configPath)
	if err != nil {
		log.Panic(err)
	}

	// init bot
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(5*time.Second))
	defer cancel()
	client := urlfetch.Client(ctx)
	bot, err := tgbotapi.NewBotAPIWithClient(botConfig.BotKey, client)
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
	http.HandleFunc("/", handle)
	appengine.Main()
}

func handle(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	client := urlfetch.Client(ctx)
	resp, err := client.Get("https://www.google.com/")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "HTTP GET returned status %v", resp.Status)
	fmt.Fprintln(w, "Hello, world!")
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

func inlineKeyboarButtons() tgbotapi.InlineKeyboardMarkup {
	replyRow := []tgbotapi.InlineKeyboardButton{}
	//like button
	likeKeyboardButton := tgbotapi.NewInlineKeyboardButtonData("like", "like")
	replyRow = append(replyRow, likeKeyboardButton)

	//unlike button
	unlikeKeyboardButton := tgbotapi.NewInlineKeyboardButtonData("unlike", "unlike")
	replyRow = append(replyRow, unlikeKeyboardButton)

	//download button
	downloadKeyboardButton := tgbotapi.NewInlineKeyboardButtonData("download", "download")
	replyRow = append(replyRow, downloadKeyboardButton)

	buttons := tgbotapi.NewInlineKeyboardMarkup(replyRow)
	return buttons
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
		buttons := inlineKeyboarButtons()
		msg.ReplyMarkup = buttons
		mybot.bot.Send(msg)
	}
}
