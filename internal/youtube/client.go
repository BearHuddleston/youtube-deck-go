package youtube

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

type Client struct {
	service *youtube.Service
}

type SearchResult struct {
	ID           string
	Title        string
	ThumbnailURL string
	Type         string
}

type VideoInfo struct {
	ID           string
	Title        string
	ThumbnailURL string
	Duration     string
	PublishedAt  time.Time
}

func New(apiKey string) (*Client, error) {
	ctx := context.Background()
	service, err := youtube.NewService(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("youtube: create service: %w", err)
	}
	return &Client{service: service}, nil
}

func (c *Client) SearchChannels(ctx context.Context, query string, maxResults int64) ([]SearchResult, error) {
	call := c.service.Search.List([]string{"snippet"}).
		Q(query).
		Type("channel").
		MaxResults(maxResults)

	resp, err := call.Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("youtube: search channels: %w", err)
	}

	results := make([]SearchResult, 0, len(resp.Items))
	for _, item := range resp.Items {
		results = append(results, SearchResult{
			ID:           item.Snippet.ChannelId,
			Title:        item.Snippet.Title,
			ThumbnailURL: item.Snippet.Thumbnails.Default.Url,
			Type:         "channel",
		})
	}
	return results, nil
}

func (c *Client) SearchPlaylists(ctx context.Context, query string, maxResults int64) ([]SearchResult, error) {
	call := c.service.Search.List([]string{"snippet"}).
		Q(query).
		Type("playlist").
		MaxResults(maxResults)

	resp, err := call.Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("youtube: search playlists: %w", err)
	}

	results := make([]SearchResult, 0, len(resp.Items))
	for _, item := range resp.Items {
		results = append(results, SearchResult{
			ID:           item.Id.PlaylistId,
			Title:        item.Snippet.Title,
			ThumbnailURL: item.Snippet.Thumbnails.Default.Url,
			Type:         "playlist",
		})
	}
	return results, nil
}

func (c *Client) FetchChannelVideos(ctx context.Context, channelID string, maxResults int64) ([]VideoInfo, error) {
	channelResp, err := c.service.Channels.List([]string{"contentDetails"}).
		Id(channelID).
		Context(ctx).
		Do()
	if err != nil {
		return nil, fmt.Errorf("youtube: get channel: %w", err)
	}
	if len(channelResp.Items) == 0 {
		return nil, fmt.Errorf("youtube: channel not found: %s", channelID)
	}

	uploadsPlaylistID := channelResp.Items[0].ContentDetails.RelatedPlaylists.Uploads
	return c.FetchPlaylistVideos(ctx, uploadsPlaylistID, maxResults)
}

func (c *Client) FetchPlaylistVideos(ctx context.Context, playlistID string, maxResults int64) ([]VideoInfo, error) {
	call := c.service.PlaylistItems.List([]string{"snippet", "contentDetails"}).
		PlaylistId(playlistID).
		MaxResults(maxResults)

	resp, err := call.Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("youtube: fetch playlist: %w", err)
	}

	videoIDs := make([]string, 0, len(resp.Items))
	for _, item := range resp.Items {
		videoIDs = append(videoIDs, item.ContentDetails.VideoId)
	}

	if len(videoIDs) == 0 {
		return nil, nil
	}

	videoResp, err := c.service.Videos.List([]string{"snippet", "contentDetails"}).
		Id(videoIDs...).
		Context(ctx).
		Do()
	if err != nil {
		return nil, fmt.Errorf("youtube: fetch video details: %w", err)
	}

	videos := make([]VideoInfo, 0, len(videoResp.Items))
	for _, v := range videoResp.Items {
		publishedAt, _ := time.Parse(time.RFC3339, v.Snippet.PublishedAt)
		thumbnailURL := ""
		if v.Snippet.Thumbnails != nil && v.Snippet.Thumbnails.Medium != nil {
			thumbnailURL = v.Snippet.Thumbnails.Medium.Url
		}
		videos = append(videos, VideoInfo{
			ID:           v.Id,
			Title:        v.Snippet.Title,
			ThumbnailURL: thumbnailURL,
			Duration:     v.ContentDetails.Duration,
			PublishedAt:  publishedAt,
		})
	}
	return videos, nil
}
