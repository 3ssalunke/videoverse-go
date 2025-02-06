package repository

import (
	"errors"
	"log"

	"github.com/3ssalunke/videoverse/db"
	"gorm.io/gorm"
)

type VideoRepository interface {
	CreateVideo(video *db.Video) error
	GetVideoByID(id string) (*db.Video, error)
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

func (r *VideoRepositoryImpl) GetVideoByID(id string) (*db.Video, error) {
	var video db.Video
	result := r.db.Where("id = ?", id).First(&video)
	if result.Error != nil {
		log.Println("[repo] error while getting video by id: ", result.Error.Error())
		return nil, errors.New("error getting video by id")
	}

	return &video, nil
}
