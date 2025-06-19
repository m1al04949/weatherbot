package app

import (
	"log/slog"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/m1al04949/weatherbot/internal/clients/openweather"
	"github.com/m1al04949/weatherbot/internal/config"
	"github.com/m1al04949/weatherbot/internal/handler"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func RunBot() error {
	// Initialize config
	cfg := config.MustLoad()
	// Initialize logger
	log := setupLogger(cfg.Env)

	log.Info("starting application",
		slog.Any("cfg", cfg))

	// Start bot
	bot, err := tgbotapi.NewBotAPI(cfg.BotToken)
	if err != nil {
		return err
	}
	bot.Debug = true

	log.Info("Authorized on account", slog.String("botname", bot.Self.UserName))

	// Initialize OpenWeather cliens
	owClient := openweather.New(cfg.OpenWeatherKey)
	// Initialize Handler
	handler := handler.New(log, bot, owClient)

	// Start listening
	handler.Start()

	return nil
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case envDev:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case envProd:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}

	return log
}
