package notification

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"
)

// SendTelegramNotification отправляет сообщение через Telegram Bot API.
func SendTelegramNotification(chatID, message string) error {
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		return fmt.Errorf("TELEGRAM_BOT_TOKEN не задан")
	}

	telegramURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", token)

	// Формирование данных для POST запроса
	data := url.Values{}
	data.Set("chat_id", chatID)
	data.Set("text", message)
	data.Set("parse_mode", "Markdown") // или "MarkdownV2", если используете другой стиль

	resp, err := http.PostForm(telegramURL, data)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ошибка отправки уведомления, статус: %d", resp.StatusCode)
	}
	return nil
}

func FormatDateRussian(t time.Time) string {
	months := []string{
		"января", "февраля", "марта", "апреля", "мая", "июня",
		"июля", "августа", "сентября", "октября", "ноября", "декабря",
	}
	day := t.Day()
	month := months[t.Month()-1] // t.Month() возвращает значение от 1 до 12
	year := t.Year()
	return fmt.Sprintf("%d %s %d", day, month, year)
}

func FormatDeadline(deadline time.Time) string {
	now := time.Now()
	daysLeft := int(deadline.Sub(now).Hours() / 24)

	var daysText string
	switch {
	case daysLeft < 0:
		return "⌛️ Срок истек"
	case daysLeft == 0:
		return "⏳ Сегодня в " + deadline.Format("15:04")
	case daysLeft == 1:
		daysText = "1 день"
	case daysLeft > 1 && daysLeft < 5:
		daysText = fmt.Sprintf("%d дня", daysLeft)
	default:
		daysText = fmt.Sprintf("%d дней", daysLeft)
	}

	return fmt.Sprintf("📅 %s (%s осталось)",
		deadline.Format("02.01.2006 в 15:04"),
		daysText,
	)
}
