package cacherepository

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/m1al04949/weatherbot/internal/cache/redis"
	"github.com/m1al04949/weatherbot/internal/clients/openweather"
	"github.com/m1al04949/weatherbot/internal/config"
	"github.com/m1al04949/weatherbot/internal/models"
)

type CacheRepository struct {
	Cfg   *config.Config
	Log   *slog.Logger
	Cache *redis.WeatherCache
}

func New(cfg *config.Config, log *slog.Logger, cache *redis.WeatherCache) *CacheRepository {
	return &CacheRepository{
		Cfg:   cfg,
		Log:   log,
		Cache: cache,
	}
}

func (cr *CacheRepository) FreshCache(ctx context.Context, log *slog.Logger, owClient *openweather.OpenWeatherClient) {
	cities := []string{"Санкт-Петербург", "Москва", "Коломна", "Орск"}
	ticker := time.NewTicker(time.Duration(cr.Cfg.TTL) * time.Minute)
	defer ticker.Stop()

	log.Info("freshing cache is started")

	var cacheWeather models.CacheWeather

	for {
		select {
		case <-ticker.C:
			for _, city := range cities {
				cacheWeather.City = city
				cord, err := owClient.Coordinates(city)
				if err != nil {
					log.Error(fmt.Sprintf("error get coordinates for %s: %s)", city, err.Error()))
					continue
				}

				cacheWeather.Lat = cord.Lat
				cacheWeather.Lon = cord.Lon
				weather, err := owClient.CurrentWeather(cord.Lat, cord.Lon)
				if err != nil {
					log.Error(fmt.Sprintf("error get weather for %s: %s)", city, err.Error()))
					continue
				}

				cacheWeather.Weather = *weather
				if err := cr.Cache.UpdateWeather(ctx, cacheWeather); err != nil {
					log.Error(fmt.Sprintf("error refresh cache for %s: %s)", city, err.Error()))
				}
			}
			log.Info("weather update data")
		case <-ctx.Done():
			return
		}
	}
}
