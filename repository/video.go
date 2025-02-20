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
	GetVideosByIDs(ids []string) ([]db.Video, error)
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

func (r *VideoRepositoryImpl) GetVideosByIDs(ids []string) ([]db.Video, error) {
	var videos []db.Video
	result := r.db.Where("id in (?)", ids).Find(&videos)
	if result.Error != nil {
		{
			log.Println("[repo] error while getting videos by ids: ", result.Error.Error())
			return nil, errors.New("error getting videos by ids")
		}
	}

	return videos, nil
}
