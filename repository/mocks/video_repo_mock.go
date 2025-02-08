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

func (m *MockVideoRepositoryImpl) GetVideoByID(id string) (*db.Video, error) {
	args := m.Called(id)

	if args.Get(0) != nil {
		return args.Get(0).(*db.Video), args.Error(1)
	}

	return nil, args.Error(1)
}
