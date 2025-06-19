package handler

import (
	"fmt"
	"log/slog"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/m1al04949/weatherbot/internal/clients/openweather"
	f "github.com/m1al04949/weatherbot/internal/lib/format"
	"github.com/m1al04949/weatherbot/internal/models"
)

type Handler struct {
	log      *slog.Logger
	bot      *tgbotapi.BotAPI
	owClient *openweather.OpenWeatherClient
}

var currentLocation models.CordinatesResponse

// Init handler
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
	// Check updates
	for update := range updates {
		h.handlerUpdate(update)
	}
}

// Processing new updates
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

	// Get Handle Mode
	if update.Message.Text == "Ввести вручную" {
		h.messageOther(update.Message.Chat.ID)
		return
	}

	// Get Forecast
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
	h.messageCurrentWeather(update)
}

// /start message
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

// other message
func (h *Handler) messageOther(id int64) {

	msg := tgbotapi.NewMessage(id, "Введите имя населенного пункта")
	msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)

	h.bot.Send(msg)
}

// current weather message
func (h *Handler) messageCurrentWeather(update tgbotapi.Update) {
	var text strings.Builder

	replyKeyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Прогноз"),
			tgbotapi.NewKeyboardButton("Назад"),
		),
	)

	weather, err := h.owClient.CurrentWeather(currentLocation.Lat, currentLocation.Lon)
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

	text.WriteString(fmt.Sprintf("Прогноз погоды в населенном пункте %s. \n \n", currentLocation.Name))
	text.WriteString(f.FormatWeatherMessage(*weather))

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, text.String())
	msg.ReplyToMessageID = update.Message.MessageID
	msg.ReplyMarkup = replyKeyboard
	h.bot.Send(msg)
}

// forecast message handler
func (h *Handler) messageForecast(update tgbotapi.Update) {

	var text strings.Builder

	todayDate := time.Now().Format(f.DateFormat)
	targetHour := 13 // on 13:00 every next day
	replyKeyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Назад"),
		),
	)

	// Get forecast
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

	// Filter forecast
	var todayForecast, nextDaysForecast []models.Weather
	processedDays := make(map[string]bool)

	for _, item := range *forecast {
		itemTime, err := time.Parse(f.DateTimeFormat, item.Date)
		if err != nil {
			h.log.Error(err.Error())
			continue
		}

		// Today
		if itemTime.Format(f.DateFormat) == todayDate {
			todayForecast = append(todayForecast, item)
			continue
		}

		// Next day on 13:00
		dateStr := itemTime.Format(f.DateFormat)
		if !processedDays[dateStr] && itemTime.Hour() >= targetHour-1 && itemTime.Hour() <= targetHour+1 {
			nextDaysForecast = append(nextDaysForecast, item)
			processedDays[dateStr] = true
		}
	}

	text.WriteString(fmt.Sprintf("Прогноз погоды в населенном пункте %s. \n \n", currentLocation.Name))

	// Today
	text.WriteString("=== Сегодня ===\n")
	for _, item := range todayForecast {
		text.WriteString(f.FormatWeatherMessage(item))
	}

	// Next days
	if len(nextDaysForecast) > 0 {
		text.WriteString("\n=== На следующие дни ===\n")
		for _, item := range nextDaysForecast {
			text.WriteString(f.FormatWeatherMessage(item))
		}
	}

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, text.String())
	msg.ReplyMarkup = replyKeyboard
	h.bot.Send(msg)
}
