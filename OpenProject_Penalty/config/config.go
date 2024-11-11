package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type AppConfig struct {
	Tgtoken     string
	Apitoken    string
	Db_driver   string
	Db_host     string
	Db_port     string
	Db_user     string
	Db_password string
	Db_name     string
}

func LoadConfig() *AppConfig {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error load config" + err.Error())
		return nil
	}
	return &AppConfig{Tgtoken: os.Getenv("API_Telegram"), Apitoken: os.Getenv("API_OpenProject"),
		Db_driver: os.Getenv("Driver"), Db_host: os.Getenv("DB_host"), Db_port: os.Getenv("DB_port"),
		Db_user: os.Getenv("DB_user"), Db_password: os.Getenv("DB_password"), Db_name: os.Getenv("DB_name")}
}
