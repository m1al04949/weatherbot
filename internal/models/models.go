package models

type Cordinates struct {
	Lat float64
	Lon float64
}

type CordinatesResponse struct {
	Name string  `json:"name"`
	Lat  float64 `json:"lat"`
	Lon  float64 `json:"lon"`
}

type Weather struct {
	Date        string
	Description string
	Temp        float64
	Humidity    int64
	Speed       float64
}

type WeatherResponse struct {
	Weather []struct {
		Description string `json:"description"`
	} `json:"weather"`
	Main struct {
		Temp     float64 `json:"temp"`
		Humidity int64   `json:"humidity"`
	} `json:"main"`
	Wind struct {
		Speed float64 `json:"speed"`
	} `json:"wind"`
}

type ForecastWeatherResponse struct {
	List []struct {
		Main struct {
			Temp     float64 `json:"temp"`
			Humidity int64   `json:"humidity"`
		} `json:"main"`
		Weather []struct {
			Description string `json:"description"`
		} `json:"weather"`
		Wind struct {
			Speed float64 `json:"speed"`
		} `json:"wind"`
		Date string `json:"dt_txt"`
	} `json:"list"`
}
