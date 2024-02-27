package thirdparties

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/bxcodec/go-clean-arch/domain"
	kkdaiYoutube "github.com/kkdai/youtube/v2"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

type YoutubeRepository struct{}

func NewYoutubeRepo() *YoutubeRepository {
	return &YoutubeRepository{}
}

func (s *YoutubeRepository) SearchVideos(text string, amount int64) ([]domain.SearchedVideo, error) {
	developerKey := os.Getenv("GOOGLE_DEV_KEY")

	query := flag.String("query", text, "Search term")
	maxResults := flag.Int64("max-results", amount, "Max YouTube results")

	flag.Parse()

	service, err := youtube.NewService(context.TODO(), option.WithAPIKey(developerKey))
	if err != nil {
		log.Printf("Error creating new YouTube client: %v", err)
		return nil, err
	}

	// Make the API call to YouTube.
	call := service.Search.List([]string{"id,snippet"}).
		Q(*query).
		MaxResults(*maxResults)
	response, err := call.Do()
	if err != nil {
		log.Printf("Error querying videos: %v", err)
		return nil, err
	}

	// Group video, channel, and playlist results in separate lists.
	videos := make(map[string]string)
	channels := make(map[string]string)
	playlists := make(map[string]string)

	// Iterate through each item and add it to the correct list.
	results := make([]domain.SearchedVideo, len(videos))
	for _, item := range response.Items {
		switch item.Id.Kind {
		case "youtube#video":
			videos[item.Id.VideoId] = item.Snippet.Title
			results = append(results, domain.SearchedVideo{
				Id:      item.Id.VideoId,
				Title:   item.Snippet.Title,
				Website: "youtube",
				URL:     fmt.Sprintf("https://www.youtube.com/watch?v=%s", item.Id.VideoId),
				Thumbnail: &domain.VideoThumbnail{
					URL:    item.Snippet.Thumbnails.Default.Url,
					Width:  item.Snippet.Thumbnails.Default.Width,
					Height: item.Snippet.Thumbnails.Default.Height,
				},
			})
		case "youtube#channel":
			channels[item.Id.ChannelId] = item.Snippet.Title
		case "youtube#playlist":
			playlists[item.Id.PlaylistId] = item.Snippet.Title
		}
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
