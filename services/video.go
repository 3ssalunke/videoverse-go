package services

import (
	"fmt"
	"log"
	"mime/multipart"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

const (
	MAX_VIDEO_SIZE_MB          = 150
	MIN_VIDEO_DURATION_SECONDS = 5
	MAX_VIDEO_DURATION_SECONDS = 50
	UPLOAD_DIR                 = "./video_store/"
)

type VideoMeta struct {
	FileSize     int64
	FileDuration float64
}

type UploadedVideo struct {
	Filename string
	FilePath string
}

func UploadVideo(file multipart.File, fileHeader *multipart.FileHeader) (*UploadedVideo, error) {
	if err := os.MkdirAll(UPLOAD_DIR, os.ModePerm); err != nil {
		log.Printf("failed to create directory: %s", err.Error())
		return nil, fmt.Errorf("failed to create directory")
	}

	filename := fmt.Sprintf("%d-%s", time.Now().UnixMilli(), fileHeader.Filename)
	savePath := UPLOAD_DIR + filename
	outputFile, err := os.Create(savePath)
	if err != nil {
		log.Printf("failed to save video file: %s", err.Error())
		return nil, fmt.Errorf("failed to save video file")
	}
	defer outputFile.Close()

	if _, err := file.Seek(0, 0); err != nil {
		log.Printf("failed to reset file pointer: %s", err.Error())
		return nil, fmt.Errorf("failed to reset file pointer")
	}
	if _, err := outputFile.ReadFrom(file); err != nil {
		log.Printf("failed to save video: %s", err.Error())
		return nil, fmt.Errorf("failed to save video")
	}

	return &UploadedVideo{Filename: filename, FilePath: savePath}, nil
}

func ValidateVideo(file multipart.File, fileHeader *multipart.FileHeader) (*VideoMeta, error) {
	maxVideoSizeBytes := MAX_VIDEO_SIZE_MB * 1024 * 1024
	if fileHeader.Size > int64(maxVideoSizeBytes) {
		log.Printf("file size exceeds the maximum allowed size of %d MB", MAX_VIDEO_SIZE_MB)
		return nil, fmt.Errorf("file size exceeds the maximum allowed size of %d MB", MAX_VIDEO_SIZE_MB)
	}

	tempFile, err := os.CreateTemp("", "upload-*.mp4")
	if err != nil {
		log.Printf("failed to create temporary file: %s", err.Error())
		return nil, fmt.Errorf("failed to create temporary file")
	}
	defer os.Remove(tempFile.Name())

	if _, err := file.Seek(0, 0); err != nil {
		log.Printf("failed to reset file pointer: %s", err.Error())
		return nil, fmt.Errorf("failed to reset file pointer")
	}
	if _, err := tempFile.ReadFrom(file); err != nil {
		log.Printf("failed to copy file to temp file: %s", err.Error())
		return nil, fmt.Errorf("failed to copy file to temp file")
	}

	duration, err := getVideoDuration(tempFile.Name())
	if err != nil {
		log.Printf("failed to get video duration: %s", err.Error())
		return nil, fmt.Errorf("failed to get video duration")
	}

	if duration < MIN_VIDEO_DURATION_SECONDS || duration > MAX_VIDEO_DURATION_SECONDS {
		log.Printf("video duration must be between %d and %d seconds", MIN_VIDEO_DURATION_SECONDS, MAX_VIDEO_DURATION_SECONDS)
		return nil, fmt.Errorf("video duration must be between %d and %d seconds", MIN_VIDEO_DURATION_SECONDS, MAX_VIDEO_DURATION_SECONDS)
	}

	return &VideoMeta{FileSize: fileHeader.Size, FileDuration: duration}, nil
}

var execCommand = exec.Command

func getVideoDuration(filePath string) (float64, error) {
	cmd := execCommand("ffprobe", "-v", "error", "-show_entries", "format=duration", "-of", "default=noprint_wrappers=1:nokey=1", filePath)
	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("error running ffprobe: %w", err)
	}

	duration, err := strconv.ParseFloat(strings.TrimSpace(string(output)), 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse duration: %w", err)
	}

	return duration, err
}
