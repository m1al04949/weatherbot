package format

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/m1al04949/weatherbot/internal/models"
)

const (
	DateTimeFormat = "2006-01-02 15:04:05"
	DateFormat     = "2006-01-02"
)

// Help function: formating message
func FormatWeatherMessage(item models.Weather) string {

	var (
		first  string
		result string
	)

	WeekdaysRu := []string{
		"Вс",
		"Пн",
		"Вт",
		"Ср",
		"Чт",
		"Пт",
		"Сб",
	}

	today := time.Now().Format(DateFormat)

	itemTime, _ := time.Parse("2006-01-02 15:04:05", item.Date)

	switch {
	case item.Date == "":
		result = fmt.Sprintf("Сегодня температура %d°C %s\n%s %s\nветер %d м/с %s",
			int(math.Round(item.Temp)), getTempEmoji(int(math.Round(item.Temp))),
			item.Description, getWeatherEmoji(item.Description),
			int(math.Round(item.Speed)), getWindEmoji(int(math.Round(item.Speed))),
		)
		return result
	case itemTime.Format(DateFormat) == today:
		first = itemTime.Format("15:04")
	default:
		first = itemTime.Format("02.01")
	}

	result = "├─────────────────────────┤\n"
	result += fmt.Sprintf("│  %5s         %3d°C %-8s    %-8s %-3dм/с \n",
		first,
		int(math.Round(item.Temp)),
		getTempEmoji(int(math.Round(item.Temp))),
		getWeatherEmoji(item.Description),
		int(math.Round(item.Speed)),
	)
	result += fmt.Sprintf("│ %5s              %20s  \n",
		WeekdaysRu[int(itemTime.Weekday())],
		item.Description,
	)

	return result
}

func getWeatherEmoji(weather string) string {
	switch {
	case strings.Contains(weather, "ясно"):
		return "☀️"
	case strings.Contains(weather, "снег"):
		return "🌨️"
	case strings.Contains(weather, "облачно с прояснениями"):
		return "🌤️"
	case strings.Contains(weather, "пасмурно"):
		return "☁️"
	case strings.Contains(weather, "небольшой дождь"):
		return "☔"
	case strings.Contains(weather, "дождь"):
		return "🌧️"
	case strings.Contains(weather, "гроза"):
		return "⛈️"
	case strings.Contains(weather, "облачность"):
		return "⛅"
	default:
		return "🌈"
	}
}

func getTempEmoji(temp int) string {
	switch {
	case temp > 25:
		return "🔥"
	case temp > 14 && temp < 26:
		return "😊"
	case temp > 9 && temp < 15:
		return "😐"
	case temp > 0 && temp < 10:
		return "🥺"
	case temp < 1:
		return "❄️"
	case temp < -9:
		return "🧊"
	default:
		return "😊"
	}
}

func getWindEmoji(speed int) string {
	switch {
	case speed > 10:
		return "💨💨"
	case speed > 5:
		return "💨"
	default:
		return "🍃"
	}
}
