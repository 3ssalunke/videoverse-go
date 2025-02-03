package mocks

import (
	"github.com/3ssalunke/videoverse/db"
	"github.com/stretchr/testify/mock"
)

type MockVideoRepositoryImpl struct {
	mock.Mock
}

func (m *MockVideoRepositoryImpl) CreateVideo(video *db.Video) error {
	args := m.Called(video)
	return args.Error(0)
}
