package handlers

import (
	"Penalty/internal/client"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Notification struct {
	ID        int    `json:"id"`
	Reason    string `json:"reason"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
	Links     struct {
		Actor struct {
			Title string `json:"title"`
		} `json:"actor"`
		Project struct {
			Title string `json:"title"`
		} `json:"project"`
		Resource struct {
			Title string `json:"title"`
		} `json:"resource"`
	} `json:"_links"`
}

type NotificationsResponse struct {
	Embedded struct {
		Elements []Notification `json:"elements"`
	} `json:"_embedded"`
}

// Функция для получения уведомлений и их форматирования
func GetNotifications(c *client.Client) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		url := fmt.Sprintf("%s/api/v3/notifications", c.BaseURL)

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		req.Header.Set("Authorization", c.CreateBasicAuthHeader())

		resp, err := c.HTTPClient.Do(req)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Ошибка при получении уведомлений: %s", resp.Status)})
			return
		}

		var notificationsResponse NotificationsResponse
		if err := json.NewDecoder(resp.Body).Decode(&notificationsResponse); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		notifications := (notificationsResponse.Embedded.Elements)
		ctx.JSON(http.StatusOK, gin.H{"notifications": notifications})
	}
}
