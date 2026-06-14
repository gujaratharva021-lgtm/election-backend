// config/database.go
package config

import (
	"election_backend/models"
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func ConnectDatabase(cfg *Config) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPass, cfg.DBName, cfg.DBSSLMode,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatal("❌ Database connection failed:", err)
	}

	log.Println("✅ Database connected!")

	// Auto Migrate — tables automatically ban jayenge
	err = db.AutoMigrate(
		&models.Election{},
		&models.Party{},
		&models.Constituency{},
		&models.Candidate{},
		&models.Result{},
		&models.VoteTrend{},
		&models.AIInsight{},
		&models.ScrapeLog{},
	)
	if err != nil {
		log.Fatal("❌ Migration failed:", err)
	}

	log.Println("✅ Database migrated!")
	DB = db
}
