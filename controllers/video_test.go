package controllers

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/3ssalunke/videoverse/db"
	repoMock "github.com/3ssalunke/videoverse/repository/mocks"
	"github.com/3ssalunke/videoverse/services"
	"github.com/3ssalunke/videoverse/utils"
	fsMock "github.com/3ssalunke/videoverse/utils/mocks"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupRouter(route, method string, handlers ...gin.HandlerFunc) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	if method == "post" {
		r.POST(route, handlers...)
	}
	return r
}

var mockRepo = new(repoMock.MockVideoRepositoryImpl)
var mockFS = new(fsMock.MockFileSystem)

// Upload video
var mockUploadVideo = func(video multipart.File, fileHeader *multipart.FileHeader) (*services.UploadedVideo, error) {
	return &services.UploadedVideo{Filename: "test.mp4", FilePath: "/mock/path/test.mp4"}, nil
}

var mockValidateVideo = func(video multipart.File, fileHeader *multipart.FileHeader) (*services.VideoMeta, error) {
	return &services.VideoMeta{FileSize: 1024, FileDuration: 30}, nil
}

func TestUploadVideo_Success(t *testing.T) {
	mockRepo.On("CreateVideo", mock.Anything).Return(nil)

	services.UploadVideo = mockUploadVideo
	services.ValidateVideo = mockValidateVideo

	videoController := NewVideoController(mockRepo, mockFS)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	fileWriter, _ := writer.CreateFormFile("video", "test.mp4")
	io.Copy(fileWriter, bytes.NewReader([]byte("mock video data")))
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	router := setupRouter("/upload", "post", videoController.UploadVideo)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"message":"video uploaded succesfully"`)
	mockRepo.AssertExpectations(t)

	t.Cleanup(func() {
		mockRepo = new(repoMock.MockVideoRepositoryImpl)
		mockFS = new(fsMock.MockFileSystem)
	})
}

func TestUploadVideo_MissingFile(t *testing.T) {
	videoController := NewVideoController(mockRepo, mockFS)

	req := httptest.NewRequest(http.MethodPost, "/upload", nil)
	w := httptest.NewRecorder()

	router := setupRouter("/upload", "post", videoController.UploadVideo)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), `"error":`)

	t.Cleanup(func() {
		mockRepo = new(repoMock.MockVideoRepositoryImpl)
		mockFS = new(fsMock.MockFileSystem)
	})
}

func TestUploadVideo_InvalidVideoFile(t *testing.T) {
	services.ValidateVideo = func(video multipart.File, fileHeader *multipart.FileHeader) (*services.VideoMeta, error) {
		return nil, errors.New("invalid video file")
	}

	videoController := NewVideoController(mockRepo, mockFS)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	fileWriter, _ := writer.CreateFormFile("video", "test.mp4")
	io.Copy(fileWriter, bytes.NewReader([]byte("mock video data")))
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	router := setupRouter("/upload", "post", videoController.UploadVideo)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), `{"error":"invalid video file"}`)

	t.Cleanup(func() {
		mockRepo = new(repoMock.MockVideoRepositoryImpl)
		mockFS = new(fsMock.MockFileSystem)
	})
}

func TestUploadVideo_UploadFailure(t *testing.T) {
	services.UploadVideo = func(video multipart.File, fileHeader *multipart.FileHeader) (*services.UploadedVideo, error) {
		return nil, errors.New("video upload failed")
	}
	services.ValidateVideo = func(video multipart.File, fileHeader *multipart.FileHeader) (*services.VideoMeta, error) {
		return &services.VideoMeta{FileSize: 1024, FileDuration: 30}, nil
	}

	videoController := NewVideoController(mockRepo, mockFS)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	fileWriter, _ := writer.CreateFormFile("video", "test.mp4")
	io.Copy(fileWriter, bytes.NewReader([]byte("mock video data")))
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	router := setupRouter("/upload", "post", videoController.UploadVideo)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), `{"error":"video upload failed"}`)

	t.Cleanup(func() {
		mockRepo = new(repoMock.MockVideoRepositoryImpl)
		mockFS = new(fsMock.MockFileSystem)
	})
}

func TestUploadVideo_DatabaseFailure(t *testing.T) {
	mockRepo.On("CreateVideo", mock.Anything).Return(errors.New("database error"))

	services.UploadVideo = func(video multipart.File, fileHeader *multipart.FileHeader) (*services.UploadedVideo, error) {
		return &services.UploadedVideo{Filename: "test.mp4", FilePath: "/mock/path/test.mp4"}, nil
	}

	services.ValidateVideo = func(video multipart.File, fileHeader *multipart.FileHeader) (*services.VideoMeta, error) {
		return &services.VideoMeta{FileSize: 1024, FileDuration: 30}, nil
	}

	videoController := NewVideoController(mockRepo, mockFS)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	fileWriter, _ := writer.CreateFormFile("video", "test.mp4")
	io.Copy(fileWriter, bytes.NewReader([]byte("mock video data")))
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	router := setupRouter("/upload", "post", videoController.UploadVideo)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), `"error":"database error"`)
	mockRepo.AssertExpectations(t)

	t.Cleanup(func() {
		mockRepo = new(repoMock.MockVideoRepositoryImpl)
		mockFS = new(fsMock.MockFileSystem)
	})
}

// Trim video
func TestTrimVideo_Success(t *testing.T) {
	mockRepo.On("GetVideoByID", "video-id").Return(&db.Video{
		ID:       "video-id",
		Name:     "test.mp4",
		Path:     "videos/test.mp4",
		Duration: 120.0,
		Size:     3000000,
	}, nil)
	mockRepo.On("CreateVideo", mock.Anything).Return(nil)

	mockFileInfo := fsMock.MockFileInfo{FileSize: 2000000}
	mockFS.On("Stat", mock.AnythingOfType("string")).Return(mockFileInfo, nil)

	services.TrimVideo = func(videoPath, outputPath string, startTs, endTs, duration float64) error {
		return nil
	}

	videoController := NewVideoController(mockRepo, mockFS)
	router := setupRouter("/trim", "post", videoController.TrimVideo)

	trimReq := utils.VideoTrimRequest{
		VideoID: "video-id",
		StartTS: 10,
		EndTS:   50,
	}
	jsonVal, _ := json.Marshal(trimReq)

	req := httptest.NewRequest(http.MethodPost, "/trim", bytes.NewBuffer(jsonVal))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"message":"video trimmed successfully"`)
	mockRepo.AssertExpectations(t)

	t.Cleanup(func() {
		mockRepo = new(repoMock.MockVideoRepositoryImpl)
		mockFS = new(fsMock.MockFileSystem)
	})
}

func TestTrimVideo_InvalideJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/trim", bytes.NewBuffer([]byte(`{invalid-json}`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	videoController := NewVideoController(mockRepo, mockFS)
	router := setupRouter("/trim", "post", videoController.TrimVideo)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), `"error"`)
	mockRepo.AssertExpectations(t)
}

func TestTrimVideo_VideoNotFound(t *testing.T) {
	mockRepo.On("GetVideoByID", "non-existent-video-id").Return(nil, errors.New("video not found"))

	videoController := NewVideoController(mockRepo, mockFS)
	router := setupRouter("/trim", "post", videoController.TrimVideo)

	trimReq := utils.VideoTrimRequest{
		VideoID: "non-existent-video-id",
		StartTS: 10,
		EndTS:   50,
	}
	jsonVal, _ := json.Marshal(trimReq)

	req := httptest.NewRequest(http.MethodPost, "/trim", bytes.NewBuffer(jsonVal))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), `"error":"video not found"`)
	mockRepo.AssertExpectations(t)

	t.Cleanup(func() {
		mockRepo = new(repoMock.MockVideoRepositoryImpl)
	})
}

func TestTrimVideo_InvalidTimestamps(t *testing.T) {
	mockRepo.On("GetVideoByID", "video-id").Return(&db.Video{
		ID:       "video-id",
		Name:     "test.mp4",
		Path:     "videos/test.mp4",
		Duration: 120.0,
		Size:     3000000,
	}, nil)

	videoController := NewVideoController(mockRepo, mockFS)
	router := setupRouter("/trim", "post", videoController.TrimVideo)

	trimReq := utils.VideoTrimRequest{
		VideoID: "video-id",
		StartTS: 130,
		EndTS:   50,
	}
	jsonVal, _ := json.Marshal(trimReq)

	req := httptest.NewRequest(http.MethodPost, "/trim", bytes.NewBuffer(jsonVal))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), `"error":"incorrect start and end timestamp bounds"`)
	mockRepo.AssertExpectations(t)

	t.Cleanup(func() {
		mockRepo = new(repoMock.MockVideoRepositoryImpl)
	})
}

func TestTrimVideo_FileNotFound(t *testing.T) {
	mockRepo.On("GetVideoByID", "video-id").Return(&db.Video{
		ID:       "video-id",
		Name:     "test.mp4",
		Path:     "videos/test.mp4",
		Duration: 120.0,
		Size:     3000000,
	}, nil)

	mockFS.On("Stat", mock.AnythingOfType("string")).Return(nil, os.ErrNotExist)

	videoController := NewVideoController(mockRepo, mockFS)
	router := setupRouter("/trim", "post", videoController.TrimVideo)

	trimReq := utils.VideoTrimRequest{
		VideoID: "video-id",
		StartTS: 10,
		EndTS:   50,
	}
	jsonVal, _ := json.Marshal(trimReq)

	req := httptest.NewRequest(http.MethodPost, "/trim", bytes.NewBuffer(jsonVal))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), `"error":"video file not found"`)
	mockRepo.AssertExpectations(t)

	t.Cleanup(func() {
		mockRepo = new(repoMock.MockVideoRepositoryImpl)
		mockFS = new(fsMock.MockFileSystem)
	})
}
