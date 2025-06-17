package handler

import (
	"fmt"
	"log/slog"
	"math"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/m1al04949/weatherbot/internal/clients/openweather"
)

type Handler struct {
	log      *slog.Logger
	bot      *tgbotapi.BotAPI
	owClient *openweather.OpenWeatherClient
}

func New(log *slog.Logger, bot *tgbotapi.BotAPI, owClient *openweather.OpenWeatherClient) *Handler {
	return &Handler{
		log:      log,
		bot:      bot,
		owClient: owClient,
	}
}

func (h *Handler) Start() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := h.bot.GetUpdatesChan(u)

	for update := range updates {
		h.handlerUpdate(update)
	}
}

func (h *Handler) handlerUpdate(update tgbotapi.Update) {
	if update.Message == nil {
		return
	}

	// If we got a message
	h.log.Info("New message",
		slog.String("username", update.Message.From.UserName),
		slog.String("message", update.Message.Text))

	// /start
	if update.Message.Text == "/start" || update.Message.Text == "Назад" {
		h.messageStart(update.Message.Chat.ID)
		return
	}

	// Handle Mode
	if update.Message.Text == "Ввести вручную" {
		h.messageOther(update.Message.Chat.ID)
		return
	}

	replyKeyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Назад"),
		),
	)

	// Get coordinates
	cord, err := h.owClient.Coordinates(update.Message.Text)
	if err != nil {
		h.log.Error(err.Error())
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Такой населенный пункт не найден")
		msg.ReplyToMessageID = update.Message.MessageID
		msg.ReplyMarkup = replyKeyboard
		h.bot.Send(msg)
		return
	}

	// Get weather
	weather, err := h.owClient.Weather(cord.Lat, cord.Lon)
	if err != nil {
		h.log.Error(err.Error())
		msg := tgbotapi.NewMessage(
			update.Message.Chat.ID,
			fmt.Sprintf("Температура в населенном пункте %s не определена", update.Message.Text))
		msg.ReplyToMessageID = update.Message.MessageID
		msg.ReplyMarkup = replyKeyboard
		h.bot.Send(msg)
		return
	}

	msg := tgbotapi.NewMessage(
		update.Message.Chat.ID,
		fmt.Sprintf("Температура в населенном пункте %s: %d °C", update.Message.Text, int(math.Round(weather.Temp))))
	msg.ReplyToMessageID = update.Message.MessageID
	msg.ReplyMarkup = replyKeyboard
	h.bot.Send(msg)
}

func (h *Handler) messageStart(id int64) {
	replyKeyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Санкт-Петербург"),
			tgbotapi.NewKeyboardButton("Ввести вручную"),
		),
	)

	msg := tgbotapi.NewMessage(id, "Узнать погоду в населенном пункте")
	msg.ReplyMarkup = replyKeyboard

	h.bot.Send(msg)
}

func (h *Handler) messageOther(id int64) {

	msg := tgbotapi.NewMessage(id, "Введите имя населенного пункта")
	msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)

	h.bot.Send(msg)
}
