package service

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5" // Импортируем библиотеку для работы с Telegram API
)

// Структуры для представления уведомлений и их компонентов
type Notification struct {
	ID        int       `json:"id"`
	Reason    string    `json:"reason"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	Links     Links     `json:"_links"`
}

type Links struct {
	Actor    Actor    `json:"actor"`
	Project  Project  `json:"project"`
	Resource Resource `json:"resource"`
}

type Actor struct {
	Title string `json:"title"`
}

type Project struct {
	Title string `json:"title"`
}

type Resource struct {
	Title string `json:"title"`
}

type NotificationResponse struct {
	Notifications []Notification `json:"notifications"`
}

// Функция для получения уведомлений
func GetNotifications() ([]Notification, error) {
	resp, err := http.Get("http://localhost:3000/notifications")
	if err != nil {
		return nil, fmt.Errorf("error fetching notifications: %w", err)
	}
	defer resp.Body.Close()

	var response NotificationResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	return response.Notifications, nil
}

// Функция для форматирования уведомлений для Telegram
func FormatNotificationsForTelegram(notifications []Notification) string {
	var formattedNotifications string
	for _, n := range notifications {
		formattedNotifications += fmt.Sprintf(
			"───────────────────────────\n"+
				"🔔 **Уведомление ID:** %d\n"+
				"👤 **Автор:** %s\n"+
				"📂 **Проект:** %s\n"+
				"📌 **Ресурс:** %s\n"+
				"🕒 **Дата создания:** %s\n"+
				"───────────────────────────\n",
			n.ID,
			n.Links.Actor.Title,
			n.Links.Project.Title,
			n.Links.Resource.Title,
			n.CreatedAt.Format("2006-01-02 15:04:05"),
		)
	}
	return formattedNotifications
}

// Функция для отправки сообщений в Telegram
func SendMessageToTelegram(bot *tgbotapi.BotAPI, chatID int64, message string) {
	msg := tgbotapi.NewMessage(chatID, message)
	msg.ParseMode = "Markdown"
	_, err := bot.Send(msg)
	if err != nil {
		log.Printf("Error sending message: %v", err)
	}
}
