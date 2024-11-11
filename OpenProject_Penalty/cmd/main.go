package main

import (
	"Penalty/config"
	"Penalty/internal/client"
	"Penalty/internal/database"
	"Penalty/internal/handlers"
	openprojectoauth "Penalty/internal/openproject_oauth"
	"Penalty/telegram"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {

	cfg := config.LoadConfig()
	db := database.ConnectDB(cfg)

	tgbot := telegram.NewBot(cfg.Tgtoken, db)
	go tgbot.StartBot()

	apiClient := client.NewClient(
		"http://146.19.183.11/dc",
		cfg.Apitoken,
	)
	go StartDailyOverdueTaskCheck(apiClient)
	go StartNotificationCheck(apiClient)

	openprojectoauth.InitVeriable()
	router := gin.Default()
	router.SetTrustedProxies(nil)
	router.GET("/callback", openprojectoauth.CallBackHandler(cfg, db))
	router.GET("/users", handlers.GetUsers(apiClient))
	router.GET("/penalty", handlers.GetTask(apiClient))

	router.GET("/notifications", handlers.GetNotifications(apiClient))

	router.Run(":3000")
}
func StartDailyOverdueTaskCheck(c *client.Client) {
	for {
		RunOverdueTaskCheck(c)

		time.Sleep(24 * time.Hour)
	}
}
func RunOverdueTaskCheck(c *client.Client) {
	tasks, err := handlers.FetchWorkPackages(c)
	if err != nil {
		fmt.Printf("Failed to fetch tasks: %v\n", err)
		return
	}

	overdueTasks := handlers.DueDateTask(tasks)
	taskDetailsList := handlers.FetchDetailsForOverdueTasks(c, overdueTasks)

	for _, task := range taskDetailsList {
		fmt.Printf("Task ID: %d, Penalty: %.2f\n", task.ID, task.Penalty)
	}
}
func StartNotificationCheck(c *client.Client) {
	for {
		handlers.GetNotifications(c)
		time.Sleep(3 * time.Minute)
		fmt.Println(handlers.GetNotifications(c))
	}

}
