package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func NewDatabase(viper *viper.Viper, log *logrus.Logger) *gorm.DB {
	username := viper.GetString("database.username")
	password := viper.GetString("database.password")
	host := viper.GetString("database.host")
	port := viper.GetInt("database.port")
	database := viper.GetString("database.name")
	idleConnection := viper.GetInt("database.pool.idle")
	maxConnection := viper.GetInt("database.pool.max")
	maxLifeTimeConnection := viper.GetInt("database.pool.lifetime")

	sslMode := viper.GetString("database.ssl_mode")
	sslConfig := ""
	if sslMode == "false" || sslMode == "disable" {
		sslConfig = "&tls=false"
	}
	
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local%s", 
		username, password, host, port, database, sslConfig)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.New(&logrusWriter{Logger: log}, logger.Config{
			SlowThreshold:             time.Second * 5,
			Colorful:                  false,
			IgnoreRecordNotFoundError: true,
			ParameterizedQueries:      true,
			LogLevel:                  logger.Info,
		}),
	})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	connection, err := db.DB()
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	connection.SetMaxIdleConns(idleConnection)
	connection.SetMaxOpenConns(maxConnection)
	connection.SetConnMaxLifetime(time.Second * time.Duration(maxLifeTimeConnection))

	// Execute SQL migrations if configured
	if viper.IsSet("database.migrations_dir") {
		migrationsDir := viper.GetString("database.migrations_dir")
		log.WithFields(logrus.Fields{
			"directory": migrationsDir,
		}).Info("Executing SQL migrations")
		
		// Run the migrations
		err = executeMigrations(db, migrationsDir, log)
		if err != nil {
			log.WithFields(logrus.Fields{
				"error": err.Error(),
			}).Fatal("Failed to execute migrations")
		}
	}

	return db
}

// executeMigrations runs all the SQL migrations from the migrations directory
func executeMigrations(db *gorm.DB, migrationsDir string, log *logrus.Logger) error {
	// Check if migrations directory exists
	_, err := os.Stat(migrationsDir)
	if os.IsNotExist(err) {
		log.WithFields(logrus.Fields{
			"directory": migrationsDir,
		}).Warn("Migrations directory does not exist")
		return nil
	}

	// Get all the up migration files
	files, err := filepath.Glob(filepath.Join(migrationsDir, "*_*.up.sql"))
	if err != nil {
		return fmt.Errorf("failed to list migration files: %w", err)
	}

	// Sort files by name to ensure correct order
	// This assumes migration files are named with timestamps like: 20250517220803_create_table_shops.up.sql
	// which will naturally sort in chronological order
	
	// Create migrations table if it doesn't exist
	err = db.Exec(`
		CREATE TABLE IF NOT EXISTS migrations (
			id varchar(255) PRIMARY KEY,
			applied_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`).Error
	if err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Execute each migration
	for _, file := range files {
		// Extract migration ID from filename
		filename := filepath.Base(file)
		parts := strings.Split(filename, "_")
		if len(parts) < 2 {
			log.WithFields(logrus.Fields{
				"file": filename,
			}).Warn("Skipping migration with invalid filename format")
			continue
		}
		
		migrationID := parts[0]

		// Check if migration has already been applied
		var count int64
		err = db.Table("migrations").Where("id = ?", migrationID).Count(&count).Error
		if err != nil {
			return fmt.Errorf("failed to check migration status: %w", err)
		}

		if count > 0 {
			log.WithFields(logrus.Fields{
				"id":   migrationID,
				"file": filename,
			}).Info("Migration already applied, skipping")
			continue
		}

		// Read migration file
		sql, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", filename, err)
		}

		// Execute migration in a transaction
		err = db.Transaction(func(tx *gorm.DB) error {
			// Apply migration
			// Split SQL statements by semicolon to execute them one by one
			statements := strings.Split(string(sql), ";")
			for _, stmt := range statements {
				// Skip empty statements
				stmt = strings.TrimSpace(stmt)
				if stmt == "" {
					continue
				}
				
				err = tx.Exec(stmt).Error
				if err != nil {
					return fmt.Errorf("failed to execute statement in migration %s: %w\nStatement: %s", filename, err, stmt)
				}
			}

			// Record migration as applied
			err = tx.Exec("INSERT INTO migrations (id) VALUES (?)", migrationID).Error
			if err != nil {
				return fmt.Errorf("failed to record migration %s: %w", filename, err)
			}

			return nil
		})

		if err != nil {
			return err
		}

		log.WithFields(logrus.Fields{
			"id":   migrationID,
			"file": filename,
		}).Info("Migration applied successfully")
	}

	return nil
}

type logrusWriter struct {
	Logger *logrus.Logger
}

func (l *logrusWriter) Printf(message string, args ...interface{}) {
	l.Logger.Tracef(message, args...)
}
