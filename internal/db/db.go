package db

import (
	"database/sql"
	"fmt"
	"tosdrgo/internal/config"
	"tosdrgo/internal/logger"

	_ "github.com/lib/pq"
)

var DB *sql.DB
var SubDB *sql.DB

type dbConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

func createConnection(conf dbConfig, readOnly bool) (*sql.DB, error) {
	options := ""
	if readOnly {
		options = " options='--default_transaction_read_only=on'"
	}

	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s%s",
		conf.Host,
		conf.Port,
		conf.User,
		conf.Password,
		conf.DBName,
		conf.SSLMode,
		options,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)

	if err = db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

func InitDB() error {
	// Initialize main database
	mainConfig := dbConfig{
		Host:     config.AppConfig.Database.Host,
		Port:     config.AppConfig.Database.Port,
		User:     config.AppConfig.Database.User,
		Password: config.AppConfig.Database.Password,
		DBName:   config.AppConfig.Database.DBName,
		SSLMode:  config.AppConfig.Database.SSLMode,
	}

	var err error
	DB, err = createConnection(mainConfig, false)
	if err != nil {
		logger.LogError(err, "Failed to open main database connection")
		return err
	}

	// Initialize submissions database
	subConfig := dbConfig{
		Host:     config.AppConfig.SubmissionsDatabase.Host,
		Port:     config.AppConfig.SubmissionsDatabase.Port,
		User:     config.AppConfig.SubmissionsDatabase.User,
		Password: config.AppConfig.SubmissionsDatabase.Password,
		DBName:   config.AppConfig.SubmissionsDatabase.DBName,
		SSLMode:  config.AppConfig.SubmissionsDatabase.SSLMode,
	}

	SubDB, err = createConnection(subConfig, false)
	if err != nil {
		logger.LogError(err, "Failed to open submissions database connection")
		CloseDB() // Close main DB if submissions DB fails
		return err
	}

	logger.LogDebug("Database connections established successfully")
	return nil
}

func CloseDB() {
	if DB != nil {
		_ = DB.Close()
	}
	if SubDB != nil {
		_ = SubDB.Close()
	}
}
