package handler

import (
	"fmt"
	"log/slog"
	"math"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/m1al04949/weatherbot/internal/clients/openweather"
	"github.com/m1al04949/weatherbot/internal/models"
)

type Handler struct {
	log      *slog.Logger
	bot      *tgbotapi.BotAPI
	owClient *openweather.OpenWeatherClient
}

var currentLocation models.CordinatesResponse

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

	// Forecast
	if update.Message.Text == "Прогноз" {
		h.messageForecast(update)
		return
	}

	replyKeyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Прогноз"),
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

	// Get Current weather
	currentLocation.Name = update.Message.Text
	currentLocation.Lat = cord.Lat
	currentLocation.Lon = cord.Lon
	weather, err := h.owClient.CurrentWeather(cord.Lat, cord.Lon)
	if err != nil {
		h.log.Error(err.Error())
		msg := tgbotapi.NewMessage(
			update.Message.Chat.ID,
			fmt.Sprintf("Погода в населенном пункте %s не определена", update.Message.Text))
		msg.ReplyToMessageID = update.Message.MessageID
		msg.ReplyMarkup = replyKeyboard
		h.bot.Send(msg)
		return
	}
	msg := tgbotapi.NewMessage(
		update.Message.Chat.ID,
		fmt.Sprintf(`Погода в населенном пункте %s.
Температура %d °C, %s.
Влажность %d%%.
Скорость ветра %d м/с.`,
			update.Message.Text,
			int(math.Round(weather.Temp)), weather.Description,
			weather.Humidity,
			int(math.Round(weather.Speed)),
		))
	msg.ReplyToMessageID = update.Message.MessageID
	msg.ReplyMarkup = replyKeyboard
	h.bot.Send(msg)
}

func (h *Handler) messageStart(id int64) {
	replyKeyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Санкт-Петербург"),
			tgbotapi.NewKeyboardButton("Москва"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Коломна"),
			tgbotapi.NewKeyboardButton("Орск"),
		),
		tgbotapi.NewKeyboardButtonRow(
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

func (h *Handler) messageForecast(update tgbotapi.Update) {

	var text string

	replyKeyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Назад"),
		),
	)

	forecast, err := h.owClient.ForecastWeather(currentLocation.Lat, currentLocation.Lon)
	if err != nil {
		h.log.Error(err.Error())
		msg := tgbotapi.NewMessage(
			update.Message.Chat.ID,
			fmt.Sprintf("Прогноз погоды в населенном пункте %s не определен", update.Message.Text))
		msg.ReplyToMessageID = update.Message.MessageID
		msg.ReplyMarkup = replyKeyboard
		h.bot.Send(msg)
		return
	}

	for _, day := range *forecast {
		if text == "" {
			text = fmt.Sprintf("Прогноз в населенном пункте %s. \n",
				currentLocation.Name)
		}

		text = text + fmt.Sprintf("%s. Температура %d °C, %s. Влажность %d%%. Скорость ветра %d м/с. \n",
			day.Date, int(math.Round(day.Temp)), day.Description, day.Humidity, int(math.Round(day.Speed)))
	}

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, text)
	msg.ReplyToMessageID = update.Message.MessageID
	msg.ReplyMarkup = replyKeyboard
	h.bot.Send(msg)
}
