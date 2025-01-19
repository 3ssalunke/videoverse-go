package db

import (
	"log"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Init() {
	var err error
	DB, err = gorm.Open(sqlite.Open("video-service.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect to database", err)
	}

	DB.AutoMigrate(&Video{}, &SharedLink{})
}
