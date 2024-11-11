package telegram

import (
	"Penalty/internal/database"
	openprojectoauth "Penalty/internal/openproject_oauth"
	"Penalty/penalty"
	service "Penalty/telegram/service"
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type User struct {
	TelegramID   string
	AccessToken  string
	RefreshToken string
	IsAuthorized bool
}

type TelegramBot struct {
	token string
	bot   *tgbotapi.BotAPI
	db    *sql.DB
}

func NewBot(newToken string, db *sql.DB) *TelegramBot {
	return &TelegramBot{token: newToken, db: db}
}

func (tgbot *TelegramBot) StartBot() {
	var err error
	tgbot.bot, err = tgbotapi.NewBotAPI(tgbot.token)
	if err != nil {
		log.Fatalf("Failed to create bot API: %v", err)
	}

	tgbot.bot.Debug = true
	log.Printf("Initialize Telegram_Bot: %s", tgbot.bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := tgbot.bot.GetUpdatesChan(u)

	for update := range updates {
		if update.CallbackQuery != nil {
			tgbot.CallBack(update, tgbot.db) // Передаем db
		} else if update.Message != nil && update.Message.IsCommand() {
			tgbot.Command(update)
		}
	}
}

// Вынесенный метод Command
func (tbot *TelegramBot) Command(update tgbotapi.Update) {
	command := update.Message.Command()
	switch command {
	case "start":
		tbot.auth(update)

	case "get_Informations":
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Здесь можно получить информацию.")
		msg.ReplyMarkup = Menu // Убедитесь, что Menu корректно определен
		tbot.sendMessage(msg)

	default:
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Неизвестная команда.")
		tbot.sendMessage(msg)
	}
}

func (tbot *TelegramBot) handleGetCommand(update tgbotapi.Update) {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID,
		"Добро пожаловать в бот OpenProject\n\n"+
			"/get_Informations - Получить информацию.\n\n"+
			"/create_Task - Создать задачу.\n\n"+
			"/delete_Task - Удалить задачу.")
	tbot.sendMessage(msg)
}

func (tbot *TelegramBot) auth(update tgbotapi.Update) {
	telegramID := update.Message.Chat.ID
	telegramIDStr := strconv.FormatInt(telegramID, 10)

	if tokens, exists := openprojectoauth.UserTokens[telegramIDStr]; exists && tokens.AccessToken != "" {
		tbot.handleGetCommand(update)
	} else {
		log.Println("Токены не найдены, отправляем ссылку для авторизации")
		authURL := "http://146.19.183.11/dc/oauth/authorize?response_type=code&client_id=50jnK5XQFJ2m-6V8_iYcVFr5s9uCzJTBlJqIyXX7nEM&redirect_uri=http://localhost:3000/callback&scope=api_v3&prompt=consent&state=" + telegramIDStr
		authMsg := tgbotapi.NewMessage(update.Message.Chat.ID, "Для авторизации в OpenProject перейдите по следующей ссылке:\n"+authURL)
		tbot.sendMessage(authMsg)
	}
}

func (tbot *TelegramBot) CallBack(update tgbotapi.Update, db *sql.DB) {
	data := update.CallbackQuery.Data
	chatID := update.CallbackQuery.From.ID
	firstName := update.CallbackQuery.From.FirstName
	lastName := update.CallbackQuery.From.LastName

	var text string

	log.Printf("Received callback query data: %s", data)

	switch data {

	case "callback_mytasks":
		telegramID := update.CallbackQuery.From.ID
		log.Printf("Получен Telegram ID: %d", telegramID)

		// Получаем OpenProject ID из базы данных
		openProjectID, err := database.GetOpenProjectIDByTelegramID(db, telegramID)
		if err != nil {
			log.Printf("Не удалось получить OpenProject ID для Telegram ID %d: %v", telegramID, err)
			text = "Ошибка при получении идентификатора OpenProject. Пожалуйста, попробуйте позже."
			break
		}

		log.Printf("Получен OpenProject ID: %d", openProjectID)

		// Теперь получаем задачи
		tasks, err := service.GetUserTasks(db, openProjectID)
		if err != nil {
			log.Printf("Не удалось получить задачи для OpenProject ID %d: %v", openProjectID, err)
			text = "Ошибка при получении задач."
		} else {
			text = service.FormatMyTasksForTelegram(tasks)
		}

	case "callback_penalties":
		log.Printf("Пользователь %s %s запросил свои задачи", firstName, lastName)
		overdueTasks, err := service.GetOverdueTasks()
		if err != nil {
			text = fmt.Sprintf("Ошибка при получении просроченных задач: %v", err)
			log.Printf("Ошибка при получении просроченных задач: %v", err)
		} else {
			if len(overdueTasks) == 0 {
				text = "Нет просроченных задач."
			} else {
				text = service.FormatOverdueTasksForTelegram(overdueTasks)
			}
		}

	case "callback_mypenalty":
		log.Printf("Пользователь %s %s запросил свои штрафы", firstName, lastName)
		dueDate := time.Now().Add(-48 * time.Hour)
		overdueDays := time.Since(dueDate).Hours() / 24
		penaltyTask := penalty.CalculatePenalty("Task", overdueDays)
		penaltyBug := penalty.CalculatePenalty("Bug", overdueDays)
		penaltyFeature := penalty.CalculatePenalty("Feature", overdueDays)
		penaltyEpic := penalty.CalculatePenalty("Epic", overdueDays)
		text = fmt.Sprintf(
			"Виды штрафов:\n\n"+
				"📝 Задача (Task): %.2f сомони\n"+
				"🐞 Баг (Bug): %.2f сомони\n"+
				"✨ Фича (Feature): %.2f сомони\n"+
				"📚 Эпик (Epic): %.2f сомони\n",
			penaltyTask, penaltyBug, penaltyFeature, penaltyEpic)

	case "callback_notifications":
		
		log.Printf("Пользователь %s %s запросил уведомления", firstName, lastName)
		notifications, err := service.GetNotifications()
		if err != nil {
			text = fmt.Sprintf("Ошибка при получении уведомлений: %v", err)
			log.Printf("Ошибка при получении уведомлений: %v", err)
		} else {
			text = service.FormatNotificationsForTelegram(notifications)
		}

	default:
		text = "Неизвестная команда."
		log.Printf("Неизвестная команда: %s", data)
	}

	msg := tgbotapi.NewMessage(chatID, text)
	tbot.sendMessage(msg)
}

func (tbot *TelegramBot) sendMessage(msg tgbotapi.Chattable) {
	if _, err := tbot.bot.Send(msg); err != nil {
		log.Printf("Error sending message: %v", err)
	}
}
