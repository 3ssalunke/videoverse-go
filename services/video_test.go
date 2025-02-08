package services

import (
	"fmt"
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

// Upload video
func TestUploadVideo(t *testing.T) {
	mockData := []byte("test video data")
	mockFile := &MockFile{data: mockData}
	fileHeader := &multipart.FileHeader{Filename: "test.mp4", Size: int64(len(mockData))}

	uploadedVideo, err := UploadVideo(mockFile, fileHeader)
	fmt.Println(uploadedVideo)
	assert.NoError(t, err)
	assert.NotNil(t, uploadedVideo)
	assert.True(t, strings.HasPrefix(uploadedVideo.FilePath, UPLOAD_DIR))

	os.Remove(uploadedVideo.FilePath)
}

// Validate Video
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

// Trim Video
func TestTrimVideo_Success(t *testing.T) {
	execCommand = func(command string, args ...string) *exec.Cmd {
		return exec.Command("cmd", "/C", "echo")
	}

	err := TrimVideo("input.mp4", "output.mp4", 10.5, 30.5, 20)
	assert.NoError(t, err)

	execCommand = exec.Command
}

func TestTrimVideo_Failure(t *testing.T) {
	execCommand = func(command string, args ...string) *exec.Cmd {
		cmd := exec.Command("echo")
		cmd.Process = &os.Process{}
		return cmd
	}

	err := TrimVideo("input.mp4", "output.mp4", 10.5, 30.5, 20)
	assert.Error(t, err)

	execCommand = exec.Command
}

// Merge video
func TestMergeVideos_Success(t *testing.T) {
	execCommand = func(command string, args ...string) *exec.Cmd {
		return exec.Command("cmd", "/C", "echo")
	}

	err := MergeVideos([]string{"input1.mp4", "input2.mp4"}, "output.mp4")
	assert.NoError(t, err)

	execCommand = exec.Command
}

func TestMergeVideos_Failure(t *testing.T) {
	execCommand = func(command string, args ...string) *exec.Cmd {
		cmd := exec.Command("echo")
		cmd.Process = &os.Process{}
		return cmd
	}

	err := MergeVideos([]string{"input1.mp4", "input2.mp4"}, "output.mp4")
	assert.Error(t, err)

	execCommand = exec.Command
}
