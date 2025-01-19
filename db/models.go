package db

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Video struct {
	ID          string       `gorm:"primary_key" json:"id"`
	Name        string       `gorm:"not null" json:"name"`
	Size        int          `gorm:"not null" json:"size"`
	Duration    int          `gorm:"not null" json:"duration"`
	Path        string       `gorm:"not null" json:"path"`
	CreatedAt   time.Time    `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	SharedLinks []SharedLink `gorm:"foreign_key:VideoID" json:"shared_links"`
}

type SharedLink struct {
	ID         string    `gorm:"primary_key" json:"id"`
	VideoID    string    `gorm:"not null" json:"video_id"`
	ExpiryDate time.Time `gorm:"not null" json:"expiry_date"`
	Video      Video     `gorm:"foreign_key:VideoID;association_foreign_key:ID" json:"video"`
}

func (v *Video) BeforeCreate(tx *gorm.DB) (err error) {
	v.ID = uuid.New().String()
	return
}

func (s *SharedLink) BeforeCreate(tx *gorm.DB) (err error) {
	s.ID = uuid.New().String()
	return
}

func (Video) TableName() string {
	return "videos"
}

func (SharedLink) TableName() string {
	return "shared_links"
}
