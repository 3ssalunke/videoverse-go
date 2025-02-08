package mocks

import (
	"os"
	"time"

	"github.com/stretchr/testify/mock"
)

type MockFileInfo struct {
	FileName string
	FileSize int64
}

func (m MockFileInfo) Name() string       { return m.FileName }
func (m MockFileInfo) Size() int64        { return m.FileSize }
func (m MockFileInfo) Mode() os.FileMode  { return 0644 }
func (m MockFileInfo) ModTime() time.Time { return time.Now() }
func (m MockFileInfo) IsDir() bool        { return false }
func (m MockFileInfo) Sys() any           { return nil }

type MockFileSystem struct {
	mock.Mock
}

func (m *MockFileSystem) Stat(name string) (os.FileInfo, error) {
	args := m.Called(name)
	if args.Get(0) != nil {
		return args.Get(0).(os.FileInfo), args.Error(1)
	}

	return nil, args.Error(1)
}
