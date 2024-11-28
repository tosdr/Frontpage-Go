package db

import (
	"database/sql"
	"fmt"
	"tosdrgo/config"
	"tosdrgo/logger"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func InitDB() error {
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s options='--default_transaction_read_only=on'",
		config.AppConfig.Database.Host,
		config.AppConfig.Database.Port,
		config.AppConfig.Database.User,
		config.AppConfig.Database.Password,
		config.AppConfig.Database.DBName,
		config.AppConfig.Database.SSLMode,
	)

	var err error
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		logger.LogError(err, "Failed to open database connection")
		return err
	}

	// Set connection pool settings
	DB.SetMaxOpenConns(25)
	DB.SetMaxIdleConns(25)

	if err = DB.Ping(); err != nil {
		logger.LogError(err, "Failed to ping database")
		return err
	}

	logger.LogDebug("Database connection established successfully")
	return nil
}

func CloseDB() {
	if DB != nil {
		DB.Close()
	}
}
