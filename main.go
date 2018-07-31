package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"strings"
	"time"

	"github.com/apex/log"
	"github.com/apex/log/handlers/cli"
	"github.com/apex/log/handlers/graylog"
	"github.com/apex/log/handlers/multi"
	"github.com/go-stack/stack"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/halink0803/telegram-unsplash-bot/common"
	"github.com/halink0803/telegram-unsplash-bot/unsplash"
)

var currentCommand string

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
	graylogHandler, _ := graylog.New("udp://127.0.0.1:12201")
	log.SetHandler(multi.New(
		cli.Default,
		graylogHandler,
	))

	if err != nil {
		log.Fatal(err.Error())
	}

	// init bot
	bot, err := tgbotapi.NewBotAPI(botConfig.BotKey)
	if err != nil {
		log.Fatal(err.Error())
	}
	bot.Debug = true

	// init unsplash client
	unsplash := unsplash.NewUnsplash(botConfig.UnsplashKey, botConfig.UnsplashSecret)

	mybot := Bot{
		bot:      bot,
		unsplash: unsplash,
	}

	for range time.Tick(time.Millisecond * 200) {
		for range time.Tick(time.Millisecond * 200) {
			log.Info("upload")
			log.Info("upload complete")
			log.Warn("upload retry")
			log.WithError(fmt.Errorf("Error %+v", stack.Caller(0))).Error("Why no line number?")
		}
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := mybot.bot.GetUpdatesChan(u)

	for update := range updates {
		mybot.handle(update)

		// msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
		// msg.ReplyToMessageID = update.Message.MessageID

		// mybot.bot.Send(msg)
	}
}

func (mybot *Bot) handle(update tgbotapi.Update) {
	if update.Message != nil && update.Message.IsCommand() {
		switch update.Message.Command() {
		case "search":
			mybot.handleSearch(update)
			break
		case "authorize":
			mybot.handleAuthorize(update)
		default:
			break
		}
	} else if update.CallbackQuery != nil {
		mybot.handleCallbackQuery(update)
	} else {
		switch currentCommand {
		case "authorize":
			mybot.handleAuthorizeCode(update)
			break
		default:
			break
		}
	}
}

func inlineKeyboarButtons(photoID string, liked bool) tgbotapi.InlineKeyboardMarkup {
	replyRow := []tgbotapi.InlineKeyboardButton{}
	if liked {
		//unlike button
		data := fmt.Sprintf("DELETE|%s", photoID)
		unlikeKeyboardButton := tgbotapi.NewInlineKeyboardButtonData("unlike", data)
		replyRow = append(replyRow, unlikeKeyboardButton)
	} else {
		//like button
		data := fmt.Sprintf("POST|%s", photoID)
		likeKeyboardButton := tgbotapi.NewInlineKeyboardButtonData("like", data)
		replyRow = append(replyRow, likeKeyboardButton)
	}

	//download button
	downloadKeyboardButton := tgbotapi.NewInlineKeyboardButtonData("download", photoID)
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
		msgContent := ""
		if photo.Description != "" {
			msgContent = fmt.Sprintf("[%s](%s)\n", photo.Description, photo.URLs.Regular)
		} else {
			msgContent = fmt.Sprintf("[%s](%s)\n", arguments, photo.URLs.Regular)
		}
		msgContent += fmt.Sprintf("Author: [%s](%s)\n", photo.User.Name, photo.User.PortfolioURL)
		msgContent += fmt.Sprintf("Likes: %d", photo.Likes)
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, msgContent)
		msg.ParseMode = "Markdown"
		buttons := inlineKeyboarButtons(photo.ID, photo.LikedByUser)
		msg.ReplyMarkup = buttons
		mybot.bot.Send(msg)
	}
}

func (mybot *Bot) handleAuthorize(update tgbotapi.Update) {
	//url to authorize user
	reqURL, err := url.Parse("https://unsplash.com")
	if err != nil {
		log.Fatal(err.Error())
	}
	reqURL.Path += "/oauth/authorize"
	params := url.Values{}
	params.Add("client_id", mybot.unsplash.UnsplashKey())
	params.Add("redirect_uri", "urn:ietf:wg:oauth:2.0:oob")
	params.Add("response_type", "code")
	reqURL.RawQuery = params.Encode()
	reqURL.RawQuery += "&scope=public+write_likes+write_followers"

	msgContent := fmt.Sprintf("Click following link to authorize then paste authorize code to the next message to authorize: %s", reqURL)
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, msgContent)
	mybot.bot.Send(msg)
	currentCommand = "authorize"
}

func (mybot *Bot) handleAuthorizeCode(update tgbotapi.Update) {
	code := update.Message.Text
	userID := update.Message.From.ID
	err := mybot.unsplash.AuthorizeUser(code, userID)
	if err != nil {
		log.Fatal(err.Error())
	}
	msgContent := fmt.Sprint("You have successfully authorize your account.")
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, msgContent)
	mybot.bot.Send(msg)
	currentCommand = ""
}

func (mybot *Bot) likePhoto(photoID string, userID int) {
	err := mybot.unsplash.LikeAPhoto(photoID, userID)
	if err != nil {
		log.Fatal(err.Error())
	}
}

func (mybot *Bot) unlikePhoto(photoID string, userID int) {
	err := mybot.unsplash.UnlikeAPhoto(photoID, userID)
	if err != nil {
		log.Fatal(err.Error())
	}
}

func (mybot *Bot) handleCallbackQuery(update tgbotapi.Update) {
	data := strings.Split(update.CallbackQuery.Data, "|")
	action := data[0]
	photoID := data[1]
	userID := update.CallbackQuery.From.ID
	liked := true
	switch action {
	case "POST":
		mybot.likePhoto(photoID, userID)
		break
	case "DELETE":
		mybot.unlikePhoto(photoID, userID)
		liked = false
		break
	}
	editButtons := inlineKeyboarButtons(photoID, liked)
	editButtonConfig := tgbotapi.NewEditMessageReplyMarkup(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, editButtons)
	mybot.bot.Send(editButtonConfig)
}
