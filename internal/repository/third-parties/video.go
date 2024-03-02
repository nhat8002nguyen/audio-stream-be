package thirdparties

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	kkdaiYoutube "github.com/kkdai/youtube/v2"
	"github.com/nhat8002nguyen/audio-stream-be/domain"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

type YoutubeRepository struct {
	query      *string
	maxResults *int64
}

func NewYoutubeRepo() *YoutubeRepository {
	r := &YoutubeRepository{
		query:      flag.String("query", "", "Search term"),
		maxResults: flag.Int64("max-results", 0, "Max YouTube results"),
	}
	// Parse flag only one, otherwise the program will be failed
	flag.Parse()
	return r
}

func (s *YoutubeRepository) SearchVideos(text string, amount int64) ([]domain.SearchedVideo, error) {
	developerKey := os.Getenv("GOOGLE_DEV_KEY")

	s.query = &text
	s.maxResults = &amount

	service, err := youtube.NewService(context.TODO(), option.WithAPIKey(developerKey))
	if err != nil {
		log.Printf("Error creating new YouTube client: %v", err)
		return nil, err
	}

	// Make the API call to YouTube to search videos by search term.
	call := service.Search.List([]string{"id, snippet"}).
		Q(*s.query).
		MaxResults(*s.maxResults)
	response, err := call.Do()
	if err != nil {
		log.Printf("Error querying videos: %v", err)
		return nil, err
	}

	// Iterate through each item and add it to the correct list.
	videoIds := make([]string, 0, amount)
	for _, item := range response.Items {
		if item.Id.Kind == "youtube#video" && item.Id.VideoId != "" {
			videoIds = append(videoIds, item.Id.VideoId)
		}
	}

	// If no videos found, return
	if len(videoIds) == 0 {
		return nil, domain.ErrNotFound
	}

	// Create a new call to get video details
	videoListCall := service.Videos.List([]string{"id", "snippet", "contentDetails"}).Id(strings.Join(videoIds, ","))
	resp, err := videoListCall.Do()
	if err != nil {
		log.Printf("Error getting video details: %v", err)
		return nil, err
	}

	results := make([]domain.SearchedVideo, 0, amount)

	// Loop through video details and access content details
	for _, item := range resp.Items {
		results = append(results, domain.SearchedVideo{
			Id:      item.Id,
			Title:   item.Snippet.Title,
			Website: "youtube",
			URL:     fmt.Sprintf("https://www.youtube.com/watch?v=%s", item.Id),
			Thumbnail: &domain.VideoThumbnail{
				URL:    item.Snippet.Thumbnails.Default.Url,
				Width:  item.Snippet.Thumbnails.Default.Width,
				Height: item.Snippet.Thumbnails.Default.Height,
			},
			Duration: item.ContentDetails.Duration,
			Channel:  item.Snippet.ChannelTitle,
		})
	}

	return results, nil
}

func (c *YoutubeRepository) GetStreamReader() (io.ReadCloser, error) {
	videoClient := kkdaiYoutube.Client{
		HTTPClient: &http.Client{},
	}
	video, err := videoClient.GetVideo("https://www.youtube.com/watch?v=shLUsd7kQCI")
	if err != nil {
		panic(err)
	}

	// Typically youtube only provides separate streams for video and audio.
	// If you want audio and video combined, take a look a the downloader package.
	formats := video.Formats.Quality("medium")
	reader, _, err := videoClient.GetStream(video, &formats[0])

	return reader, err
}
