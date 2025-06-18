package openweather

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/m1al04949/weatherbot/internal/models"
)

type OpenWeatherClient struct {
	apiKey string
}

func New(apiKey string) *OpenWeatherClient {
	return &OpenWeatherClient{
		apiKey: apiKey,
	}
}

func (o *OpenWeatherClient) Coordinates(city string) (*models.Cordinates, error) {
	op := "clients.openwather.coordinates"
	url := "http://api.openweathermap.org/geo/1.0/direct?q=%s&limit=5&appid=%s"

	resp, err := http.Get(fmt.Sprintf(url, city, o.apiKey))
	if err != nil {
		return &models.Cordinates{}, fmt.Errorf("error get coordinates in %s: %w", op, err)
	}

	if resp.StatusCode != http.StatusOK {
		return &models.Cordinates{}, fmt.Errorf("error bad status in %s: %d", op, resp.StatusCode)
	}

	var cordinatesResp []models.CordinatesResponse

	err = json.NewDecoder(resp.Body).Decode(&cordinatesResp)
	if err != nil {
		return &models.Cordinates{}, fmt.Errorf("error unmarshal response in %s: %w", op, err)
	}

	if len(cordinatesResp) == 0 {
		return &models.Cordinates{}, fmt.Errorf("error empty coordinates in %s", op)
	}

	return &models.Cordinates{
		Lat: cordinatesResp[0].Lat,
		Lon: cordinatesResp[0].Lon,
	}, nil
}

func (o *OpenWeatherClient) CurrentWeather(lat, lon float64) (*models.Weather, error) {
	op := "clients.openwather.currentweather"
	url := "https://api.openweathermap.org/data/2.5/weather?lat=%f&lon=%f&appid=%s&units=metric&lang=ru"

	resp, err := http.Get(fmt.Sprintf(url, lat, lon, o.apiKey))
	if err != nil {
		return &models.Weather{}, fmt.Errorf("error get current weather in %s: %w", op, err)
	}

	if resp.StatusCode != http.StatusOK {
		return &models.Weather{}, fmt.Errorf("error bad status in %s: %d", op, resp.StatusCode)
	}

	var weatherResp models.WeatherResponse

	err = json.NewDecoder(resp.Body).Decode(&weatherResp)
	if err != nil {
		return &models.Weather{}, fmt.Errorf("error unmarshal response in %s: %w", op, err)
	}

	return &models.Weather{
		Description: weatherResp.Weather[0].Description,
		Temp:        weatherResp.Main.Temp,
		Humidity:    weatherResp.Main.Humidity,
		Speed:       weatherResp.Wind.Speed,
	}, nil
}

func (o *OpenWeatherClient) ForecastWeather(lat, lon float64) (*[]models.Weather, error) {
	op := "clients.openwather.forecastweather"
	url := "https://api.openweathermap.org/data/2.5/forecast?lat=%f&lon=%f&appid=%s&units=metric&lang=ru"

	resp, err := http.Get(fmt.Sprintf(url, lat, lon, o.apiKey))
	if err != nil {
		return &[]models.Weather{}, fmt.Errorf("error get forecast weather in %s: %w", op, err)
	}

	if resp.StatusCode != http.StatusOK {
		return &[]models.Weather{}, fmt.Errorf("error bad status in %s: %d", op, resp.StatusCode)
	}

	var (
		forecastWeatherResp models.ForecastWeatherResponse
		forecastWeather     []models.Weather
	)

	err = json.NewDecoder(resp.Body).Decode(&forecastWeatherResp)
	if err != nil {
		return &[]models.Weather{}, fmt.Errorf("error unmarshal response in %s: %w", op, err)
	}

	for _, item := range forecastWeatherResp.List {
		weather := models.Weather{
			Date:        item.Date, // или item.DtTxt, в зависимости от вашего API
			Description: "",
			Temp:        item.Main.Temp,
			Humidity:    item.Main.Humidity,
			Speed:       item.Wind.Speed,
		}

		// Берем первое описание погоды (если массив weather не пустой)
		if len(item.Weather) > 0 {
			weather.Description = item.Weather[0].Description
		}

		forecastWeather = append(forecastWeather, weather)
	}

	return &forecastWeather, nil
}
