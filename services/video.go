package services

import (
	"bytes"
	"fmt"
	"log"
	"mime/multipart"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	MAX_VIDEO_SIZE_MB          = 150
	MIN_VIDEO_DURATION_SECONDS = 5
	MAX_VIDEO_DURATION_SECONDS = 50
	UPLOAD_DIR                 = "video_store"
)

type VideoMeta struct {
	FileSize     int64
	FileDuration float64
}

type UploadedVideo struct {
	Filename string
	FilePath string
}

var UploadVideo = func(file multipart.File, fileHeader *multipart.FileHeader) (*UploadedVideo, error) {
	if err := os.MkdirAll(UPLOAD_DIR, os.ModePerm); err != nil {
		log.Printf("[service] failed to create directory: %s", err.Error())
		return nil, fmt.Errorf("failed to create directory")
	}

	filename := fmt.Sprintf("%d-%s", time.Now().UnixMilli(), fileHeader.Filename)
	savePath := filepath.Join(UPLOAD_DIR, filename)
	outputFile, err := os.Create(savePath)
	if err != nil {
		log.Printf("[service] failed to save video file: %s", err.Error())
		return nil, fmt.Errorf("failed to save video file")
	}
	defer outputFile.Close()

	if _, err := file.Seek(0, 0); err != nil {
		log.Printf("[service] failed to reset file pointer: %s", err.Error())
		return nil, fmt.Errorf("failed to reset file pointer")
	}
	if _, err := outputFile.ReadFrom(file); err != nil {
		log.Printf("[service] failed to save video: %s", err.Error())
		return nil, fmt.Errorf("failed to save video")
	}

	return &UploadedVideo{Filename: filename, FilePath: savePath}, nil
}

var ValidateVideo = func(file multipart.File, fileHeader *multipart.FileHeader) (*VideoMeta, error) {
	maxVideoSizeBytes := MAX_VIDEO_SIZE_MB * 1024 * 1024
	if fileHeader.Size > int64(maxVideoSizeBytes) {
		log.Printf("[service] file size exceeds the maximum allowed size of %d MB", MAX_VIDEO_SIZE_MB)
		return nil, fmt.Errorf("file size exceeds the maximum allowed size of %d MB", MAX_VIDEO_SIZE_MB)
	}

	tempFile, err := os.CreateTemp("", "upload-*.mp4")
	if err != nil {
		log.Printf("[service] failed to create temporary file: %s", err.Error())
		return nil, fmt.Errorf("failed to create temporary file")
	}
	defer os.Remove(tempFile.Name())

	if _, err := file.Seek(0, 0); err != nil {
		log.Printf("[service] failed to reset file pointer: %s", err.Error())
		return nil, fmt.Errorf("failed to reset file pointer")
	}
	if _, err := tempFile.ReadFrom(file); err != nil {
		log.Printf("[service] failed to copy file to temp file: %s", err.Error())
		return nil, fmt.Errorf("failed to copy file to temp file")
	}

	duration, err := getVideoDuration(tempFile.Name())
	if err != nil {
		log.Printf("[service] failed to get video duration: %s", err.Error())
		return nil, fmt.Errorf("failed to get video duration")
	}

	if duration < MIN_VIDEO_DURATION_SECONDS || duration > MAX_VIDEO_DURATION_SECONDS {
		log.Printf("[service] video duration must be between %d and %d seconds", MIN_VIDEO_DURATION_SECONDS, MAX_VIDEO_DURATION_SECONDS)
		return nil, fmt.Errorf("video duration must be between %d and %d seconds", MIN_VIDEO_DURATION_SECONDS, MAX_VIDEO_DURATION_SECONDS)
	}

	return &VideoMeta{FileSize: fileHeader.Size, FileDuration: duration}, nil
}

var execCommand = exec.Command

func getVideoDuration(filePath string) (float64, error) {
	cmd := execCommand("ffprobe", "-v", "error", "-show_entries", "format=duration", "-of", "default=noprint_wrappers=1:nokey=1", filePath)

	var outBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &outBuf

	if err := cmd.Run(); err != nil {
		return 0, fmt.Errorf("error running ffprobe: %s %s", outBuf.String(), err.Error())
	}

	duration, err := strconv.ParseFloat(strings.TrimSpace(outBuf.String()), 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse duration: %w", err)
	}

	return duration, err
}

var TrimVideo = func(videoPath, outputPath string, startTs, endTs, duration float64) error {
	cmd := execCommand("ffmpeg",
		"-i", videoPath,
		"-ss", fmt.Sprintf("%.2f", startTs),
		"-to", fmt.Sprintf("%.2f", endTs),
		"-c", "copy", outputPath)

	var outBuf bytes.Buffer
	cmd.Stderr = &outBuf

	if err := cmd.Run(); err != nil {
		log.Printf("[service] failed to trim video: %s %s", outBuf.String(), err.Error())
		return err
	}

	return nil
}

var MergeVideos = func(videoPaths []string, outputPath string) error {
	videoPathsFilename := "videos.txt"

	videoPathsFile, err := os.Create(videoPathsFilename)
	if err != nil {
		log.Printf("[service] failed to create a file for input video paths: %s", err.Error())
		return err
	}
	defer func() {
		videoPathsFile.Close()
		os.Remove(videoPathsFilename)
	}()

	for _, path := range videoPaths {
		abspath := filepath.ToSlash(path)
		_, err := videoPathsFile.WriteString(fmt.Sprintf("file %s\n", abspath))
		if err != nil {
			log.Printf("[service] failed to write to videoPathsFile file: %s", err.Error())
			return err
		}
	}

	args := []string{"-f", "concat", "-safe", "0", "-i", videoPathsFilename, "-c", "copy", outputPath}
	cmd := execCommand("ffmpeg", args...)

	var outBuf bytes.Buffer
	cmd.Stderr = &outBuf

	if err := cmd.Run(); err != nil {
		log.Printf("[service] failed to merge videos: %s %s", outBuf.String(), err.Error())
		return err
	}

	return nil
}
