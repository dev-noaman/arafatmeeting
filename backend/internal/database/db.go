package database

import (
	"fmt"
	"log"
	"os"
	"time"

	"mini-meeting/internal/config"
	"mini-meeting/internal/models"

	"github.com/golang-migrate/migrate/v4"
	migratepostgres "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

var DB *gorm.DB

func Connect(cfg *config.DatabaseConfig) error {
	var err error

	dsn := cfg.DSN()

	gormCfg := &gorm.Config{}
	if cfg.IsCloud() {
		newLogger := gormlogger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags),
			gormlogger.Config{
				SlowThreshold:             1 * time.Second,
				LogLevel:                  gormlogger.Warn,
				IgnoreRecordNotFoundError: true,
				Colorful:                  true,
			},
		)
		gormCfg.Logger = newLogger
	}

	DB, err = gorm.Open(postgres.Open(dsn), gormCfg)
	if err != nil {
		return fmt.Errorf("failed to connect to PostgreSQL database: %w", err)
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	if cfg.IsCloud() {
		sqlDB.SetMaxOpenConns(10)
		sqlDB.SetMaxIdleConns(10)
		sqlDB.SetConnMaxLifetime(5 * time.Minute)
		sqlDB.SetConnMaxIdleTime(1 * time.Minute)
	}

	log.Println("PostgreSQL database connection established")

	return nil
}

func Migrate() error {
	log.Println("Running database migrations...")

	// Only migrate summarizer-related tables (users, meetings managed by InsForge)
	// Drop FK constraint on user_id that was created by old integer migration
	// (GORM wants varchar(36) but migration 000008 created it as INTEGER REFERENCES users(id))
	DB.Exec("ALTER TABLE summarizer_sessions DROP CONSTRAINT IF EXISTS summarizer_sessions_user_id_fkey")
	db := DB.Session(&gorm.Session{})
	err := db.AutoMigrate(
		&models.SummarizerSession{},
		&models.AudioChunk{},
		&models.Transcript{},
	)
	if err != nil {
		return fmt.Errorf("failed to run auto-migrate: %w", err)
	}

	// Then run SQL migrations
	if err := runSQLMigrations(); err != nil {
		log.Printf("Warning: SQL migrations failed: %v", err)
	}

	log.Println("Database migrations completed")
	return nil
}

// Apply migration files from the migrations folder
func runSQLMigrations() error {
	// Get the underlying sql.DB from GORM
	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get database connection: %w", err)
	}

	// Create PostgreSQL migration driver instance
	driver, err := migratepostgres.WithInstance(sqlDB, &migratepostgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create PostgreSQL migration driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance("file://migrations", "postgres", driver)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	// Run migrations
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Println("SQL migrations applied successfully")
	return nil
}

func GetDB() *gorm.DB {
	return DB
}
