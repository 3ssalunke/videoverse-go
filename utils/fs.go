package utils

import "os"

type FileSystem interface {
	Stat(name string) (os.FileInfo, error)
}

type OSFileSystem struct{}

func (OSFileSystem) Stat(name string) (os.FileInfo, error) {
	return os.Stat(name)
}
