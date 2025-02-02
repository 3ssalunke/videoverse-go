package services

import (
	"io"
	"mime/multipart"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type MockFile struct {
	data   []byte
	cursor int64
}

func (m *MockFile) Read(p []byte) (n int, err error) {
	if m.cursor >= int64(len(m.data)) {
		return 0, io.EOF
	}

	n = copy(p, m.data[m.cursor:])
	m.cursor += int64(n)
	return n, nil
}

func (m *MockFile) Seek(offset int64, whence int) (int64, error) {
	if whence == io.SeekStart {
		m.cursor = offset
	} else if whence == io.SeekCurrent {
		m.cursor += offset
	} else if whence == io.SeekEnd {
		m.cursor = int64(len(m.data)) + offset
	}

	return m.cursor, nil
}

func (m *MockFile) Close() error {
	return nil
}

func (m *MockFile) ReadAt(p []byte, off int64) (n int, err error) {
	return 0, nil
}

func TestUploadVideo(t *testing.T) {
	mockData := []byte("test video data")
	mockFile := &MockFile{data: mockData}
	fileHeader := &multipart.FileHeader{Filename: "test.mp4", Size: int64(len(mockData))}

	uploadedVideo, err := UploadVideo(mockFile, fileHeader)
	assert.NoError(t, err)
	assert.NotNil(t, uploadedVideo)
	assert.True(t, strings.HasPrefix(uploadedVideo.FilePath, UPLOAD_DIR))

	os.Remove(uploadedVideo.FilePath)
}

func TestValidateVideo_SizeExceeds(t *testing.T) {
	mockData := make([]byte, MAX_VIDEO_SIZE_MB*1024*1024+1)
	mockFile := &MockFile{data: mockData}
	fileHeader := &multipart.FileHeader{Filename: "large.mp4", Size: int64(len(mockData))}

	videoMeta, err := ValidateVideo(mockFile, fileHeader)
	assert.Error(t, err)
	assert.Nil(t, videoMeta)
}

func TestGetVideoDuration(t *testing.T) {
	execCommand = func(command string, args ...string) *exec.Cmd {
		return exec.Command("cmd", "/C", "echo 30.5")
	}

	duration, err := getVideoDuration("test.mp4")
	assert.NoError(t, err)
	assert.Equal(t, 30.5, duration)

	execCommand = exec.Command
}

func TestValidateVideo_InvalidDuration(t *testing.T) {
	execCommand = func(command string, args ...string) *exec.Cmd {
		return exec.Command("cmd", "/C", "echo 2.5")
	}

	mockData := []byte("test video data")
	mockFile := &MockFile{data: mockData}
	fileHeader := &multipart.FileHeader{Filename: "large.mp4", Size: int64(len(mockData))}

	videoMeta, err := ValidateVideo(mockFile, fileHeader)
	assert.Error(t, err)
	assert.Nil(t, videoMeta)

	execCommand = exec.Command
}
