package format

import (
	"fmt"
	"math"
	"time"

	"github.com/m1al04949/weatherbot/internal/models"
)

const (
	DateTimeFormat = "2006-01-02 15:04:05"
	DateFormat     = "2006-01-02"
)

// Help function: formating message
func FormatWeatherMessage(item models.Weather) string {

	var result string

	today := time.Now().Format(DateFormat)

	itemTime, _ := time.Parse("2006-01-02 15:04:05", item.Date)

	switch {
	case item.Date == "":
		result = "Сегодня"
	case itemTime.Format(DateFormat) == today:
		result = fmt.Sprintf("в %s", itemTime.Format("15:04"))
	default:
		result = itemTime.Format("02.01")
	}

	result = fmt.Sprintf("%s:\nТемпература %d°C, %s.\nВлажность %d%%, ветер %d м/с.\n\n",
		result,
		int(math.Round(item.Temp)),
		item.Description,
		item.Humidity,
		int(math.Round(item.Speed)),
	)

	return result
}
