package handlers

import (
	"Penalty/internal/client"
	"Penalty/penalty"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type WorkPackage struct {
	ID        int    `json:"id"`
	Subject   string `json:"subject"`
	StartDate string `json:"startDate"`
	DueDate   string `json:"dueDate"`
	Embedded  struct {
		Type struct {
			Name string `json:"name"`
		} `json:"type"`
		Responsible struct {
			Name string `json:"name"`
		} `json:"responsible"`
		Assignee struct {
			Name string `json:"name"`
		} `json:"assignee"`
	} `json:"_embedded"`
}

type TaskDetails struct {
	ID          int
	Subject     string
	Type        string
	Responsible string
	DueDate     time.Time
	Penalty     float64
}

func DueDateTask(tasks []WorkPackage) map[int]WorkPackage {
	overDueDateTasksMap := make(map[int]WorkPackage)
	nowDate := time.Now()

	for _, wp := range tasks {
		if wp.DueDate == "" {
			continue
		}

		duedate, err := time.Parse("2006-01-02", wp.DueDate)
		if err != nil {
			continue // Пропустить, если не удалось разобрать дату
		}

		if duedate.Before(nowDate) {
			overDueDateTasksMap[wp.ID] = wp
		}
	}

	return overDueDateTasksMap
}

func GetTaskDetailsByID(c *client.Client, id int) (*WorkPackage, error) {
	url := fmt.Sprintf("%s/api/v3/work_packages/%d", c.BaseURL, id)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", c.CreateBasicAuthHeader())

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch work package ID %d: %s", id, res.Status)
	}

	var wp WorkPackage
	if err := json.NewDecoder(res.Body).Decode(&wp); err != nil {
		return nil, err
	}

	return &wp, nil
}

func FetchDetailsForOverdueTasks(c *client.Client, overdueTasks map[int]WorkPackage) []TaskDetails {
	var taskDetailsList []TaskDetails

	for id, wp := range overdueTasks {
		// Получаем детали задачи по ID
		taskDetail, err := GetTaskDetailsByID(c, id)
		if err != nil {
			continue
		}

		taskType := taskDetail.Embedded.Type.Name
		responsible := taskDetail.Embedded.Assignee.Name

		// Парсим дату завершения
		duedate, err := time.Parse("2006-01-02", wp.DueDate)
		if err != nil {
			continue // Пропустить, если не удалось разобрать дату
		}

		// Рассчитываем количество просроченных дней
		overdueDays := int(time.Since(duedate).Hours() / 24)
		if overdueDays <= 0 {
			continue // Пропустить, если задача не просрочена
		}

		// Вызываем функцию CalculatePenalty
		penaltyAmount := penalty.CalculatePenalty(taskType, float64(overdueDays))

		// Создаем структуру TaskDetails
		taskDetails := TaskDetails{
			ID:          taskDetail.ID,
			Subject:     taskDetail.Subject,
			Type:        taskType,
			Responsible: responsible,
			DueDate:     duedate,
			Penalty:     penaltyAmount,
		}

		// Добавляем детали задачи в список
		taskDetailsList = append(taskDetailsList, taskDetails)
	}

	return taskDetailsList
}

func GetTask(apiClient *client.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Получаем все рабочие пакеты
		tasks, err := FetchWorkPackages(apiClient)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tasks"})
			return
		}

		// Фильтруем просроченные задачи
		overdueTasks := DueDateTask(tasks)

		// Получаем детали для просроченных задач
		taskDetailsList := FetchDetailsForOverdueTasks(apiClient, overdueTasks)

		c.JSON(http.StatusOK, taskDetailsList)
	}
}

func FetchWorkPackages(c *client.Client) ([]WorkPackage, error) {
	req, err := http.NewRequest("GET", c.BaseURL+"/api/v3/work_packages", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", c.CreateBasicAuthHeader())

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error: %s", resp.Status)
	}

	var result struct {
		Embedded struct {
			Elements []WorkPackage `json:"elements"`
		} `json:"_embedded"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Embedded.Elements, nil
}
