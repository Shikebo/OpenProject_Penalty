package database

import (
	"Penalty/config"
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

func ConnectDB(cfg *config.AppConfig) *sql.DB {
	driver := cfg.Db_driver

	dbHost := cfg.Db_host
	dbPort := cfg.Db_port
	dbUser := cfg.Db_user
	dbPassword := cfg.Db_password
	dbName := cfg.Db_name

	var db *sql.DB
	var connect string
	var err error

	if driver == "postgres" {
		connect = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			dbHost, dbPort, dbUser, dbPassword, dbName)
		db, err = sql.Open(driver, connect)
		if err != nil {
			log.Fatal("Failed to connect to PostgreSQL database:", err)
		}
	} else if driver == "mssql" {
		connect = fmt.Sprintf("server=%s;user id=%s;password=%s;port=%s;encrypt=disable;database=%s",
			dbHost, dbUser, dbPassword, dbPort, dbName)
		db, err = sql.Open(driver, connect)
		if err != nil {
			log.Fatal("Failed to connect to MSSQL database:", err)
		}
	} else {
		log.Fatal("Unsupported database driver:", driver)
	}

	err = db.Ping()
	if err != nil {
		log.Fatal("Error establishing a connection to the database:", err)
	}

	log.Println("Successfully connected to the database!")
	return db
}

func Save_ID(db *sql.DB, telegramID string, openprojectID string) error {
	query := `INSERT INTO user_ID (telegram_ID, openproject_ID) VALUES ($1, $2)`
	_, err := db.Exec(query, telegramID, openprojectID)
	if err != nil {
		return fmt.Errorf("failed to insert user ID: %w", err)
	}
	return nil
}
func GetOpenProjectIDByTelegramID(db *sql.DB, telegramID int64) (string, error) {
	var openProjectID string
	query := "SELECT openproject_id FROM users WHERE telegram_id = ?"

	err := db.QueryRow(query, telegramID).Scan(&openProjectID)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("пользователь с Telegram ID %d не найден", telegramID)
		}
		return "", fmt.Errorf("ошибка выполнения запроса: %v", err)
	}

	return openProjectID, nil
}
