package handlers

import (
	"Penalty/internal/client"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type User struct {
	ID        int    `json:"id"`        // Уникальный идентификатор пользователя
	Login     string `json:"login"`     // Логин пользователя
	FirstName string `json:"firstName"` // Имя пользователя
	LastName  string `json:"lastName"`  // Фамилия пользователя
	Email     string `json:"email"`     // Электронная почта пользователя
	Status    string `json:"status"`    // Статус пользователя (например, active, invited)
	Language  string `json:"language"`  // Язык пользователя
	CreatedAt string `json:"createdAt"` // Дата создания пользователя
	UpdatedAt string `json:"updatedAt"` // Дата последнего обновления пользователя
}

func GetUsers(c *client.Client) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req, err := http.NewRequest("GET", c.BaseURL+"/api/v3/users", nil)
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
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Errorf("error: %s", resp.Status).Error()})
			return
		}

		var result struct {
			Embedded struct {
				Elements []User `json:"elements"`
			} `json:"_embedded"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		ctx.JSON(http.StatusOK, result.Embedded.Elements)
	}
}
