package controllers

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/3ssalunke/videoverse/db"
	"github.com/3ssalunke/videoverse/repository"
	"github.com/3ssalunke/videoverse/services"
	"github.com/3ssalunke/videoverse/utils"
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
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	videoMeta, err := services.ValidateVideo(file, header)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	uploadedFile, err := services.UploadVideo(file, header)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer file.Close()

	video := &db.Video{
		Name:     uploadedFile.Filename,
		Path:     uploadedFile.FilePath,
		Duration: videoMeta.FileDuration,
		Size:     videoMeta.FileSize,
	}

	err = v.videoRepo.CreateVideo(video)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "video uploaded succesfully", "video_id": video.ID})
}

func (v *VideoController) TrimVideo(c *gin.Context) {
	var trimReqPayload utils.VideoTrimRequest

	if err := c.ShouldBindJSON(&trimReqPayload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	video, err := v.videoRepo.GetVideoByID(trimReqPayload.VideoID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if trimReqPayload.StartTS < 0 || trimReqPayload.EndTS <= 0 ||
		trimReqPayload.StartTS >= video.Duration || trimReqPayload.EndTS > video.Duration ||
		trimReqPayload.StartTS >= trimReqPayload.EndTS {
		errMessage := "incorrect start and end timestamp bounds"
		log.Println("[controller]", errMessage)
		c.JSON(http.StatusBadRequest, gin.H{"error": errMessage})
		return
	}

	_, err = os.Stat(video.Path)
	if os.IsNotExist(err) {
		errMessage := "video file not found"
		log.Println("[controller]", errMessage)
		c.JSON(http.StatusInternalServerError, gin.H{"error": errMessage})
		return
	}

	currTimestamp := time.Now().UnixMilli()
	parts := strings.Split(video.Name, "-")
	trimmedFilename := fmt.Sprintf("%d-%s", currTimestamp, strings.Join(parts[1:], "-"))
	outputPath := filepath.Join(services.UPLOAD_DIR, trimmedFilename)

	if err := services.TrimVideo(video.Path, outputPath, trimReqPayload.StartTS, trimReqPayload.EndTS, video.Duration); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to trim video"})
		return
	}

	fileInfo, err := os.Stat(outputPath)
	if err != nil {
		errMessage := "failed to get trimmed video stat"
		log.Println("[controller]", errMessage)
		c.JSON(http.StatusInternalServerError, gin.H{"error": errMessage})
		return
	}

	video = &db.Video{
		Name:     trimmedFilename,
		Path:     outputPath,
		Duration: trimReqPayload.EndTS - trimReqPayload.StartTS,
		Size:     fileInfo.Size(),
	}

	err = v.videoRepo.CreateVideo(video)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "video trimmed successfully", "video_id": video.ID})
}
