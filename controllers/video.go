package controllers

import (
	"net/http"

	"github.com/3ssalunke/videoverse/db"
	"github.com/3ssalunke/videoverse/repository"
	"github.com/3ssalunke/videoverse/services"
	"github.com/gin-gonic/gin"
)

type VideoController struct {
	videoRepo repository.VideoRepository
}

func NewVideoController(videoRepo repository.VideoRepository) *VideoController {
	return &VideoController{videoRepo}
}

func (v *VideoController) UploadVideo(c *gin.Context) {
	file, header, err := c.Request.FormFile("video")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	videoMeta, err := services.ValidateVideo(file, header)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	uploadedFile, err := services.UploadVideo(file, header)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	defer file.Close()

	video := &db.Video{
		Name:     uploadedFile.Filename,
		Path:     uploadedFile.FilePath,
		Duration: int(videoMeta.FileDuration),
		Size:     int(videoMeta.FileSize),
	}

	err = v.videoRepo.CreateVideo(video)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "video uploaded succesfully"})
}
