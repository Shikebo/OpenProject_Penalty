package openprojectoauth

import (
	"Penalty/config"
	"Penalty/internal/database"
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

var UserTokens = make(map[string]UserToken)

type UserToken struct {
	AccessToken   string
	RefreshToken  string
	TelegramID    string
	OpenProjectID string
	IsAuthorized  bool
}

var (
	clientID     string
	clientSecret string
	redirectURI  string
	tokenURL     string
)

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	Scope        string `json:"scope"`
}

func InitVeriable() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	clientID = os.Getenv("CLIENT_ID")
	clientSecret = os.Getenv("CLIENT_SECRET")
	redirectURI = os.Getenv("REDIRECT_URL")
	tokenURL = os.Getenv("TOKEN_URL")
}
func CallBackHandler(cfg *config.AppConfig, db *sql.DB) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		CallBack(ctx, cfg, db)
	}
}

func CallBack(ctx *gin.Context, cfg *config.AppConfig, db *sql.DB) {
	code := ctx.Query("code")
	telegramID := ctx.Query("state")

	if code == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Authorization code not found"})
		return
	}

	log.Println("Authorization code received:", code)

	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("client_id", clientID)
	data.Set("client_secret", clientSecret)
	data.Set("code", code)
	data.Set("redirect_uri", redirectURI)

	req, err := http.NewRequest("POST", tokenURL, bytes.NewBufferString(data.Encode()))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
		log.Println("Error creating request:", err)
		return
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send request"})
		log.Println("Error sending request:", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get access token"})
		log.Println("Failed to get access token, status code:", resp.StatusCode)
		return
	}

	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse response"})
		log.Println("Error parsing response:", err)
		return
	}

	// Сохраняем токены в UserTokens
	UserTokens[telegramID] = UserToken{
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
		TelegramID:   telegramID,
		IsAuthorized: true,
	}

	// Получаем OpenProjectID
	openProjectID, err := getOpenProjectID(tokenResp.AccessToken)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get OpenProject ID"})
		log.Println("Error getting OpenProject ID:", err)
		return
	}

	// Сохраняем идентификаторы в базу данных
	err = database.Save_ID(db, telegramID, openProjectID)
	if err != nil {
		log.Fatal("Error saving user ID:", err)
		return
	}

	log.Printf("Saved Telegram ID: %s, OpenProject ID: %s", telegramID, openProjectID)

	// Перенаправляем пользователя на Telegram с токеном
	http.Redirect(ctx.Writer, ctx.Request, "https://t.me/dctasks_bot?start="+tokenResp.AccessToken, http.StatusFound)
}

func getOpenProjectID(accessToken string) (string, error) {
	req, err := http.NewRequest("GET", "http://146.19.183.11/dc/api/v3/users/me", nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get OpenProject ID, status code: %d", resp.StatusCode)
	}

	var user struct {
		ID int `json:"id"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	return strconv.Itoa(user.ID), nil
}
