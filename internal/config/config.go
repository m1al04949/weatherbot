package config

import (
	"log"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	Env            string `yaml:"env" env:"ENV" end-default:"local"`
	BotToken       string `yaml:"bottoken" env-required:"true"`
	OpenWeatherKey string `yaml:"openweatherkey" env-required:"true"`
	HuggingFaceKey string `yaml:"huggingfacekey" env-required:"true"`
	WebhookURL     string `yaml:"webhookurl"`
	Port           string `yaml:"port"`
	Cache          `yaml:"cache"`
	Broker         `yaml:"broker"`
}

type Cache struct {
	Address  string `yaml:"address" env-required:"true"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
	TTL      int    `yaml:"ttl" env-required:"true"`
}

type Broker struct {
	Addrs   []string `yaml:"addrs" env-required:"true"`
	Retry   int      `yaml:"retry"`
	Timeout int      `yaml:"timeout"`
}

func MustLoad() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("error loading .env file")
	}

	configPath := os.Getenv("CFG_PATH")
	if configPath == "" {
		log.Fatal("config path is not set")
	}

	//check if file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("config file does not exist: %s", configPath)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("cannot read config: %s", err)
	}

	return &cfg
}
