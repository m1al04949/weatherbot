package app

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/m1al04949/weatherbot/internal/cache/redis"
	"github.com/m1al04949/weatherbot/internal/clients/huggingface"
	"github.com/m1al04949/weatherbot/internal/clients/openweather"
	"github.com/m1al04949/weatherbot/internal/config"
	"github.com/m1al04949/weatherbot/internal/handler"
	"github.com/m1al04949/weatherbot/internal/repositories/cacherepository"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func RunBot() error {
	var wg sync.WaitGroup

	// Get context
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	// Initialize config
	cfg := config.MustLoad()
	// Initialize logger
	log := setupLogger(cfg.Env)

	log.Info("starting application", slog.Any("cfg", cfg))

	// Start bot
	bot, err := tgbotapi.NewBotAPI(cfg.BotToken)
	if err != nil {
		return err
	}
	bot.Debug = true

	log.Info("authorized on account", slog.String("botname", bot.Self.UserName))

	// Initialize OpenWeather client
	owClient := openweather.New(cfg.OpenWeatherKey)
	// Initialize Hugging Face client
	hfClient := huggingface.New(cfg.HuggingFaceKey)
	// Initialize Cache
	cache := redis.NewCache(
		cfg.Cache.Address, cfg.Cache.Password, cfg.Cache.DB,
		time.Duration(cfg.Cache.TTL)*time.Minute, log)
	// Initialize repositories
	cacheRep := cacherepository.New(cfg, log, cache)
	// Freshing cache
	wg.Add(1)
	go func() {
		defer wg.Done()
		cacheRep.FreshCache(ctx, log, owClient)
	}()
	// Initialize Handler
	handler := handler.New(log, bot, owClient, hfClient, cache)

	// Start listening telegram messages
	wg.Add(1)
	go func() {
		defer wg.Done()
		handler.Start(ctx)
	}()

	// Graceful shutdown
	<-ctx.Done()
	log.Info("shutting down...")

	cache.Close()

	wg.Wait()
	log.Info("shutdown complete")

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
