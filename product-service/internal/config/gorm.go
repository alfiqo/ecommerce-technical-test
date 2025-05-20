package config

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func NewDatabase(config *viper.Viper, log *logrus.Logger) *gorm.DB {
	username := config.GetString("database.username")
	password := config.GetString("database.password")
	host := config.GetString("database.host")
	port := config.GetInt("database.port")
	dbname := config.GetString("database.name")
	
	// Get SSL and additional params configuration
	sslEnabled := config.GetBool("database.ssl")
	params := config.GetString("database.params")
	
	// Build the base connection string
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", username, password, host, port, dbname)
	
	// Add query parameters
	queryParams := "charset=utf8mb4&parseTime=True&loc=Local"
	
	// Add SSL settings
	if !sslEnabled {
		queryParams += "&tls=false"
	}
	
	// Add additional params if specified
	if params != "" {
		queryParams += "&" + params
	}
	
	// Complete the DSN
	dsn = fmt.Sprintf("%s?%s", dsn, queryParams)
	
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Failed to get database connection: %v", err)
	}
	
	sqlDB.SetMaxIdleConns(config.GetInt("database.pool.idle"))
	sqlDB.SetMaxOpenConns(config.GetInt("database.pool.open"))
	sqlDB.SetConnMaxLifetime(time.Duration(config.GetInt("database.pool.lifetime")) * time.Minute)
	
	log.Info("Database connection established")
	
	return db
}