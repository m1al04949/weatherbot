package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/m1al04949/weatherbot/internal/models"
)

type WeatherCache struct {
	client *redis.Client
	ttl    time.Duration
}

func NewCache(addr, password string, db int, ttl time.Duration) *WeatherCache {
	return &WeatherCache{
		client: redis.NewClient(&redis.Options{
			Addr:     addr,
			Password: password,
			DB:       db,
		}),
		ttl: ttl,
	}
}

// Get weather from cache
func (c *WeatherCache) GetWeather(ctx context.Context, city string) (*models.CacheWeather, error) {
	op := "redis.getweather"

	data, err := c.client.Get(ctx, cityKey(city)).Bytes()
	if err == redis.Nil {
		return nil, fmt.Errorf("%s: %s", op, "key is not exists")
	}
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	var weather models.CacheWeather

	if err := json.Unmarshal(data, &weather); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &weather, nil
}

// Update weatger in cache
func (c *WeatherCache) UpdateWeather(ctx context.Context, weather models.CacheWeather) error {
	op := "redis.updateweather"

	weather.UpdatedAt = time.Now()

	data, err := json.Marshal(weather)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if err := c.client.Set(ctx, cityKey(weather.City), data, c.ttl).Err(); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func cityKey(city string) string {
	return fmt.Sprintf("weather:%s", city)
}
