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
	fs        utils.FileSystem
}

func NewVideoController(videoRepo repository.VideoRepository, fs utils.FileSystem) *VideoController {
	return &VideoController{videoRepo, fs}
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

	_, err = v.fs.Stat(video.Path)
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

	fileInfo, err := v.fs.Stat(outputPath)
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

func (v *VideoController) MergeVideos(c *gin.Context) {
	var mergeReqPayload utils.VideosMergeRequest

	if err := c.ShouldBindJSON(&mergeReqPayload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(mergeReqPayload.VideoIDs) < 2 {
		errMessage := "please give at least two video ids in request"
		log.Println("[controller]", errMessage)
		c.JSON(http.StatusBadRequest, gin.H{"error": errMessage})
		return
	}

	videos, err := v.videoRepo.GetVideosByIDs(mergeReqPayload.VideoIDs)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(videos) != len(mergeReqPayload.VideoIDs) {
		errMessage := "one or more video IDs do not exist"
		log.Println("[controller]", errMessage)
		c.JSON(http.StatusBadRequest, gin.H{"error": errMessage})
		return
	}

	var videoFilepaths []string
	mergedVideoDuration := 0.0

	for _, video := range videos {
		_, err = v.fs.Stat(video.Path)
		if os.IsNotExist(err) {
			errMessage := fmt.Sprintf("video file not found for id %s", video.ID)
			log.Println("[controller]", errMessage)
			c.JSON(http.StatusInternalServerError, gin.H{"error": errMessage})
			return
		}
		videoFilepaths = append(videoFilepaths, video.Path)
		mergedVideoDuration += video.Duration
	}

	currTimestamp := time.Now().UnixMilli()
	mergedFilename := fmt.Sprintf("%d-merged.mp4", currTimestamp)
	outputPath := filepath.Join(services.UPLOAD_DIR, mergedFilename)

	if err := services.MergeVideos(videoFilepaths, outputPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to merge videos"})
		return
	}

	fileInfo, err := v.fs.Stat(outputPath)
	if err != nil {
		errMessage := "failed to get merged video stat"
		log.Println("[controller]", errMessage)
		c.JSON(http.StatusInternalServerError, gin.H{"error": errMessage})
		return
	}

	video := &db.Video{
		Name:     mergedFilename,
		Path:     outputPath,
		Duration: mergedVideoDuration,
		Size:     fileInfo.Size(),
	}

	err = v.videoRepo.CreateVideo(video)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "videos merged successfully", "video_id": video.ID})
}
