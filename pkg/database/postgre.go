package database

import (
	"fmt"
	"time"

	"github.com/dewadityasanjaya/card-authorization-service/config"
	"github.com/dewadityasanjaya/card-authorization-service/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

var DB *gorm.DB

func Connect(cfg *config.DatabaseConfig) *gorm.DB {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		cfg.Host,
		cfg.User,
		cfg.Password,
		cfg.Name,
		cfg.Port,
		cfg.SSLMode,
	)

	// Set GORM log level based on environment
	gormLog := gormlogger.Default.LogMode(gormlogger.Info)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormLog,
	})
	if err != nil {
		logger.Fatal("Failed to connect to database",
			zap.String("host", cfg.Host),
			zap.String("name", cfg.Name),
			zap.Error(err),
		)
	}

	// Get the underlying sql.DB to configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		logger.Fatal("Failed to get sql.DB from GORM",
			zap.Error(err),
		)
	}

	// Connection pool settings
	sqlDB.SetMaxOpenConns(25)                 // max open connections
	sqlDB.SetMaxIdleConns(10)                 // max idle connections kept in pool
	sqlDB.SetConnMaxLifetime(5 * time.Minute) // max lifetime of a connection

	// Ping to verify connection is alive
	if err := sqlDB.Ping(); err != nil {
		logger.Fatal("Database ping failed",
			zap.Error(err),
		)
	}

	logger.Info("Database connected successfully",
		zap.String("host", cfg.Host),
		zap.String("name", cfg.Name),
	)

	DB = db
	return db
}

// Close gracefully closes the database connection
func Close(db *gorm.DB) {
	sqlDB, err := db.DB()
	if err != nil {
		logger.Error("Failed to get sql.DB for closing", zap.Error(err))
		return
	}

	if err := sqlDB.Close(); err != nil {
		logger.Error("Failed to close database connection", zap.Error(err))
		return
	}

	logger.Info("Database connection closed")
}

// TxManager abstracts DB transactions so they can be mocked in tests
type TxManager interface {
	Transaction(fc func(tx *gorm.DB) error) error
}

// GormTxManager is the real implementation using *gorm.DB
type GormTxManager struct {
	db *gorm.DB
}

func NewTxManager(db *gorm.DB) TxManager {
	return &GormTxManager{db: db}
}

func (m *GormTxManager) Transaction(fc func(tx *gorm.DB) error) error {
	return m.db.Transaction(fc)
}
