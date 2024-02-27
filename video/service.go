package video

import (
	"context"
	"io"

	"github.com/nhat8002nguyen/audio-stream-be/domain"
)

type VideoRepository interface {
	SearchVideos(text string, amount int64) ([]domain.SearchedVideo, error)
	GetStreamReader() (io.ReadCloser, error)
}

type Service struct {
	videoRepo VideoRepository
}

// NewService will create a new article service object
func NewService(v VideoRepository) *Service {
	return &Service{
		videoRepo: v,
	}
}

func (s *Service) SearchVideos(ctx context.Context, text string, amount int64) ([]domain.SearchedVideo, error) {
	return s.videoRepo.SearchVideos(text, amount)
}

func (s *Service) GetStreamReader(ctx context.Context) (io.ReadCloser, error) {
	return s.videoRepo.GetStreamReader()
}
