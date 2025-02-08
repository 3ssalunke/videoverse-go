package controllers

import (
	"bytes"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/3ssalunke/videoverse/repository/mocks"
	"github.com/3ssalunke/videoverse/services"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var mockUploadVideo = func(video multipart.File, fileHeader *multipart.FileHeader) (*services.UploadedVideo, error) {
	return &services.UploadedVideo{Filename: "test.mp4", FilePath: "/mock/path/test.mp4"}, nil
}

var mockValidateVideo = func(video multipart.File, fileHeader *multipart.FileHeader) (*services.VideoMeta, error) {
	return &services.VideoMeta{FileSize: 1024, FileDuration: 30}, nil
}

func TestUploadVideo_Success(t *testing.T) {
	mockRepo := new(mocks.MockVideoRepositoryImpl)

	mockRepo.On("CreateVideo", mock.Anything).Return(nil)

	services.UploadVideo = mockUploadVideo
	services.ValidateVideo = mockValidateVideo

	videoController := NewVideoController(mockRepo)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	fileWriter, _ := writer.CreateFormFile("video", "test.mp4")
	io.Copy(fileWriter, bytes.NewReader([]byte("mock video data")))
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/upload", videoController.UploadVideo)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"message":"video uploaded succesfully"`)
	mockRepo.AssertExpectations(t)
}

func TestUploadVideo_MissingFile(t *testing.T) {
	mockRepo := new(mocks.MockVideoRepositoryImpl)
	videoController := NewVideoController(mockRepo)

	req := httptest.NewRequest(http.MethodPost, "/upload", nil)
	w := httptest.NewRecorder()

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/upload", videoController.UploadVideo)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), `"error":`)
}

func TestUploadVideo_InvalidVideoFile(t *testing.T) {
	mockRepo := new(mocks.MockVideoRepositoryImpl)

	services.ValidateVideo = func(video multipart.File, fileHeader *multipart.FileHeader) (*services.VideoMeta, error) {
		return nil, errors.New("invalid video file")
	}

	videoController := NewVideoController(mockRepo)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	fileWriter, _ := writer.CreateFormFile("video", "test.mp4")
	io.Copy(fileWriter, bytes.NewReader([]byte("mock video data")))
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/upload", videoController.UploadVideo)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), `{"error":"invalid video file"}`)
}

func TestUploadVideo_UploadFailure(t *testing.T) {
	mockRepo := new(mocks.MockVideoRepositoryImpl)

	services.UploadVideo = func(video multipart.File, fileHeader *multipart.FileHeader) (*services.UploadedVideo, error) {
		return nil, errors.New("video upload failed")
	}
	services.ValidateVideo = func(video multipart.File, fileHeader *multipart.FileHeader) (*services.VideoMeta, error) {
		return &services.VideoMeta{FileSize: 1024, FileDuration: 30}, nil
	}

	videoController := NewVideoController(mockRepo)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	fileWriter, _ := writer.CreateFormFile("video", "test.mp4")
	io.Copy(fileWriter, bytes.NewReader([]byte("mock video data")))
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/upload", videoController.UploadVideo)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), `{"error":"video upload failed"}`)
}

func TestUploadVideo_DatabaseFailure(t *testing.T) {
	mockRepo := new(mocks.MockVideoRepositoryImpl)

	mockRepo.On("CreateVideo", mock.Anything).Return(errors.New("database error"))

	services.UploadVideo = func(video multipart.File, fileHeader *multipart.FileHeader) (*services.UploadedVideo, error) {
		return &services.UploadedVideo{Filename: "test.mp4", FilePath: "/mock/path/test.mp4"}, nil
	}

	services.ValidateVideo = func(video multipart.File, fileHeader *multipart.FileHeader) (*services.VideoMeta, error) {
		return &services.VideoMeta{FileSize: 1024, FileDuration: 30}, nil
	}

	videoController := NewVideoController(mockRepo)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	fileWriter, _ := writer.CreateFormFile("video", "test.mp4")
	io.Copy(fileWriter, bytes.NewReader([]byte("mock video data")))
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/upload", videoController.UploadVideo)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), `"error":"database error"`)
	mockRepo.AssertExpectations(t)
}
