package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"

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

func (mybot *Bot) handleAuthorize(update tgbotapi.Update) {
	//url to authorize user
	reqURL, err := url.Parse("https://unsplash.com")
	if err != nil {
		log.Panic(err)
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
		log.Panic(err)
	}
	msgContent := fmt.Sprint("You have successfully authorize your account.")
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, msgContent)
	mybot.bot.Send(msg)
}

func (mybot *Bot) handleCallbackQuery(update tgbotapi.Update) {

}
