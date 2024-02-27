package domain

type SearchedVideo struct {
	Id        string          `json:"id"`
	Title     string          `json:"title"`
	Website   string          `json:"website"`
	URL       string          `json:"url"`
	Thumbnail *VideoThumbnail `json:"thumbnail"`
	Duration  string          `json:"duration"`
}

type VideoThumbnail struct {
	URL    string `json:"url"`
	Width  int64  `json:"width"`
	Height int64  `json:"height"`
}
