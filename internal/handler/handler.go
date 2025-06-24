package handler

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/m1al04949/weatherbot/internal/cache/redis"
	"github.com/m1al04949/weatherbot/internal/clients/huggingface"
	"github.com/m1al04949/weatherbot/internal/clients/openweather"
	f "github.com/m1al04949/weatherbot/internal/lib/format"
	"github.com/m1al04949/weatherbot/internal/models"
)

type Handler struct {
	log      *slog.Logger
	bot      *tgbotapi.BotAPI
	owClient *openweather.OpenWeatherClient
	hfClient *huggingface.HuggingFaceClient
	cache    *redis.WeatherCache
}

var currentLocation models.CordinatesResponse

// Init handler
func New(
	log *slog.Logger, bot *tgbotapi.BotAPI,
	owClient *openweather.OpenWeatherClient,
	hfClient *huggingface.HuggingFaceClient,
	cache *redis.WeatherCache) *Handler {
	return &Handler{
		log:      log,
		bot:      bot,
		owClient: owClient,
		hfClient: hfClient,
		cache:    cache,
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

	// Get current weather
	var (
		weather       *models.Weather
		text          strings.Builder
		replyKeyboard tgbotapi.ReplyKeyboardMarkup
	)
	// From cache
	cacheWeather, err := h.cache.GetWeather(context.Background(), update.Message.Text)
	if err != nil {
		h.log.Error(err.Error())
		// Request current weather
		weather, err = h.messageCurrentWeather(&text, update)
		if err != nil {
			h.log.Error(err.Error())
			replyKeyboard = tgbotapi.NewReplyKeyboard(
				tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButton("Назад"),
				),
			)
		}
	}
	if cacheWeather != nil {
		currentLocation.Name = cacheWeather.City
		currentLocation.Lat = cacheWeather.Lat
		currentLocation.Lon = cacheWeather.Lon
		weather = &cacheWeather.Weather
		h.log.Info("weather for from cache")
		fmt.Println(cacheWeather)
	}
	if weather != nil {
		text.WriteString(fmt.Sprintf("Прогноз погоды в населенном пункте %s. \n \n", currentLocation.Name))
		text.WriteString(f.FormatWeatherMessage(*weather))
		replyKeyboard = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("Прогноз"),
				tgbotapi.NewKeyboardButton("Назад"),
			),
		)
	}

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, text.String())
	msg.ReplyToMessageID = update.Message.MessageID
	msg.ReplyMarkup = replyKeyboard
	h.bot.Send(msg)
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
func (h *Handler) messageCurrentWeather(text *strings.Builder, update tgbotapi.Update) (*models.Weather, error) {
	cord, err := h.owClient.Coordinates(update.Message.Text)
	if err != nil {
		text.WriteString("Такой населенный пункт не найден")
		return nil, err
	}

	weather, err := h.owClient.CurrentWeather(cord.Lat, cord.Lon)
	if err != nil {
		text.WriteString(fmt.Sprintf("Погода в населенном пункте %s не определена", update.Message.Text))
		return nil, err
	}

	currentLocation.Name = update.Message.Text
	currentLocation.Lat = cord.Lat
	currentLocation.Lon = cord.Lon

	return weather, nil
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
		if itemTime.Format(f.DateFormat) == todayDate && itemTime.Hour() > time.Now().Hour() {
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

	// Formating New forecast
	text.WriteString(fmt.Sprintf("Прогноз погоды в населенном пункте %s. \n \n", currentLocation.Name))

	// Today
	text.WriteString("============= Сегодня ============\n")
	text.WriteString("┌─────────────────────────┐\n")
	text.WriteString("│ Время   Температура   Погода   Ветер \n")
	for _, item := range todayForecast {
		text.WriteString(f.FormatWeatherMessage(item))
	}
	text.WriteString("└─────────────────────────┘")

	// Next days
	if len(nextDaysForecast) > 0 {
		text.WriteString("\n======== На следующие дни ========\n")
		text.WriteString("┌─────────────────────────┐\n")
		text.WriteString("│ День    Температура   Погода    Ветер \n")
		for _, item := range nextDaysForecast {
			text.WriteString(f.FormatWeatherMessage(item))
		}
		text.WriteString("└─────────────────────────┘")
	}

	// Generate image
	// image, err := h.hfClient.GenerateWithHuggingFace(text.String())
	// if err != nil {
	// 	h.log.Error(fmt.Sprintf("failed to generate image: %s", err.Error()))
	// }
	// photo := tgbotapi.FileBytes{
	// 	Name:  "weather_forecast.png",
	// 	Bytes: image,
	// }
	// msg := tgbotapi.NewPhoto(update.Message.Chat.ID, photo)
	// msg.Caption = "Прогноз погоды 🌤️"

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, text.String())
	msg.ReplyMarkup = replyKeyboard
	h.bot.Send(msg)
}
