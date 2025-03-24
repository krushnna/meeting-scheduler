package initializers

import (
	"log"

	"github.com/krushnna/meeting-scheduler/config"
	"github.com/krushnna/meeting-scheduler/models"
	"gorm.io/gorm"
)

// InitDB initializes the database connection and performs migrations.
func InitDB() *gorm.DB {
	db, err := config.InitDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Auto migrate your models
	err = db.AutoMigrate(
		&models.Event{},
		&models.TimeSlot{},
		&models.User{},
		&models.UserAvailability{},
	)
	if err != nil {
		log.Fatalf("AutoMigrate failed: %v", err)
	}

	return db
}
