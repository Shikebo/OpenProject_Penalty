package service

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var httpClient = &http.Client{
	Timeout: 10 * time.Second,
}

type User struct {
	OpenProjectID int
	TelegramID    int64
}

type Task struct {
	ID            int        `json:"id"`
	Subject       string     `json:"subject"`
	DueDate       time.Time  `json:"dueDate"`
	CompletedAt   *time.Time `json:"completedAt,omitempty"`
	CreatedAt     time.Time  `json:"createdAt"`
	UpdatedAt     time.Time  `json:"updatedAt"`
	StartDate     time.Time  `json:"startDate,omitempty"`
	EstimatedTime float64    `json:"estimatedTime,omitempty"`
}

func (t *Task) UnmarshalJSON(data []byte) error {
	type Alias Task
	aux := &struct {
		DueDate     string  `json:"dueDate"`
		CompletedAt *string `json:"completedAt"`
		CreatedAt   string  `json:"createdAt"`
		UpdatedAt   string  `json:"updatedAt"`
		StartDate   string  `json:"startDate"`
		*Alias
	}{
		Alias: (*Alias)(t),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	var err error
	t.DueDate, err = parseDate(aux.DueDate)
	if err != nil {
		return err
	}

	if aux.CompletedAt != nil && *aux.CompletedAt != "" {
		parsedCompletedAt, err := parseDate(*aux.CompletedAt)
		if err != nil {
			return err
		}
		t.CompletedAt = &parsedCompletedAt
	}

	t.CreatedAt, err = parseDate(aux.CreatedAt)
	if err != nil {
		return err
	}
	t.UpdatedAt, err = parseDate(aux.UpdatedAt)
	if err != nil {
		return err
	}
	t.StartDate, err = parseDate(aux.StartDate)
	if err != nil {
		return err
	}
	return nil
}

func parseDate(dateStr string) (time.Time, error) {
	if dateStr == "" {
		return time.Time{}, nil
	}

	parsedDate, err := time.Parse(time.RFC3339, dateStr)
	if err == nil {
		return parsedDate, nil
	}

	parsedDate, err = time.Parse("2006-01-02", dateStr)
	if err != nil {
		return time.Time{}, err
	}
	return parsedDate, nil
}

type TaskResponse struct {
	Type     string `json:"_type"`
	Count    int    `json:"count"`
	Embedded struct {
		Elements []Task `json:"elements"`
	} `json:"_embedded"`
}

const BaseURL = "http://146.19.183.11/dc"

// Получение OpenProject ID по Telegram ID
func GetOpenProjectIDByTelegramID(db *sql.DB, telegramID int) (int, error) {
	log.Printf("Получение OpenProject ID для Telegram ID: %d", telegramID)
	query := "SELECT openproject_id FROM user_id WHERE telegram_id = $1"

	var openprojectID int

	err := db.QueryRow(query, telegramID).Scan(&openprojectID)

	if err != nil {
		log.Printf("Ошибка при выполнении SQL-запроса: %v", err)
		return 0, fmt.Errorf("Ошибка при получении идентификатора OpenProject. Пожалуйста, попробуйте позже.")
	}

	log.Printf("OpenProject ID для Telegram ID %d успешно получен: %d", telegramID, openprojectID)
	return openprojectID, nil
}

// Получение задач пользователя
func GetUserTasks(db *sql.DB, telegramIDStr string) ([]Task, error) {
	telegramID, err := strconv.Atoi(telegramIDStr)
	if err != nil {
		return nil, fmt.Errorf("ошибка преобразования Telegram ID: %v", err)
	}

	openprojectID, err := GetOpenProjectIDByTelegramID(db, telegramID)
	if err != nil {
		return nil, fmt.Errorf("не удалось получить OpenProject ID для Telegram ID %d: %v", telegramID, err)
	}

	// Формирование URL для запроса
	url := fmt.Sprintf("%s/api/v3/work_packages?filters=[{\"operator\":\"=\",\"field\":\"responsible\",\"value\":\"%d\"}]", BaseURL, openprojectID)

	// Создание нового запроса
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания запроса: %v", err)
	}

	// Установка заголовков
	apiKey := "e92d0e2ae7ec01f8d50e0c311c7c9400577d99e9081764224b56ed77469bbba4"
	credentials := fmt.Sprintf("apikey:%s", apiKey)
	encodedCredentials := base64.StdEncoding.EncodeToString([]byte(credentials))
	req.Header.Set("Authorization", "Basic "+encodedCredentials)

	// Выполнение запроса
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка выполнения запроса: %v", err)
	}
	defer resp.Body.Close()

	// Проверка статуса ответа
	if resp.StatusCode != http.StatusOK {
		body, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			return nil, fmt.Errorf("ошибка чтения тела ответа: %v", readErr)
		}
		return nil, fmt.Errorf("не удалось получить задачи: %d - %s", resp.StatusCode, string(body))
	}

	// Чтение и декодирование ответа
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения тела ответа: %v", err)
	}

	log.Printf("Response Body: %s", string(body))

	var taskResponse TaskResponse
	if err := json.Unmarshal(body, &taskResponse); err != nil {
		return nil, fmt.Errorf("ошибка декодирования ответа: %v", err)
	}

	return taskResponse.Embedded.Elements, nil
}

// Форматирование задач для Telegram
func FormatMyTasksForTelegram(tasks []Task) string {
	var formattedTasks strings.Builder
	for _, t := range tasks {
		formattedTasks.WriteString(formatMyTask(t))
	}
	return formattedTasks.String()
}

// Форматирование одной задачи
func formatMyTask(t Task) string {
	return fmt.Sprintf(
		"───────────────────────────\n"+
			"📋 **Задача ID:** %d\n"+
			"🔖 **Тема:** %s\n"+
			"📅 **Срок:** %s\n"+
			"✅ **Выполнена:** %s\n"+
			"🕒 **Дата создания:** %s\n"+
			"🕒 **Дата обновления:** %s\n"+
			"🕒 **Дата начала:** %s\n"+
			"🕒 **Оценочное время:** %.2f часов\n"+
			"───────────────────────────\n",
		t.ID,
		t.Subject,
		t.DueDate.Format("2006-01-02"),
		formatTime(t.CompletedAt),
		t.CreatedAt.Format("2006-01-02 15:04:05"),
		t.UpdatedAt.Format("2006-01-02 15:04:05"),
		t.StartDate.Format("2006-01-02 15:04:05"),
		t.EstimatedTime,
	)
}

// Форматирование времени выполнения задачи
func formatTime(t *time.Time) string {
	if t == nil || t.IsZero() {
		return "Не выполнена"
	}
	return t.Format("2006-01-02 15:04:05")
}
