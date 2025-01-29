package db

import (
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Init() {
	var err error
	// DB, err = gorm.Open(sqlite.Open("video-service.db"), &gorm.Config{})
	dsn := "host=localhost user=TEST password=TEST dbname=videoverse port=5432 sslmode=disable TimeZone=Asia/Shanghai"
	DB, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect to database", err)
	}

	DB.AutoMigrate(&Video{}, &SharedLink{})
}
