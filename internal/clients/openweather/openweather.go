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

func (o *OpenWeatherClient) Weather(lat, lon float64) (*models.Weather, error) {
	op := "clients.openwather.weather"
	url := "https://api.openweathermap.org/data/2.5/weather?lat=%f&lon=%f&appid=%s&units=metric"

	resp, err := http.Get(fmt.Sprintf(url, lat, lon, o.apiKey))
	if err != nil {
		return &models.Weather{}, fmt.Errorf("error get weather in %s: %w", op, err)
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
		Temp: weatherResp.Main.Temp,
	}, nil
}
