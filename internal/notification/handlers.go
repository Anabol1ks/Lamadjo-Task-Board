package notification

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
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
