package repository

import (
	"github.com/3ssalunke/videoverse/db"
	"gorm.io/gorm"
)

type VideoRepository interface {
	CreateVideo(video *db.Video) error
	// GetVideoByID(id string) (*db.Video, error)
	// GetAllVideos() ([]db.Video, error)
	// DeleteVideo(id string) error
}

type VideoRepositoryImpl struct {
	db *gorm.DB
}

func NewVideoRepository(db *gorm.DB) VideoRepository {
	return &VideoRepositoryImpl{db}
}

func (r *VideoRepositoryImpl) CreateVideo(video *db.Video) error {
	return r.db.Create(video).Error
}
