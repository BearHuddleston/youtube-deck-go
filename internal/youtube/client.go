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

type FetchResult struct {
	Videos        []VideoInfo
	NextPageToken string
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
			ThumbnailURL: getBestThumbnail(item.Snippet.Thumbnails),
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
			ThumbnailURL: getBestThumbnail(item.Snippet.Thumbnails),
			Type:         "playlist",
		})
	}
	return results, nil
}

func (c *Client) FetchChannelVideos(ctx context.Context, channelID string, maxResults int64) ([]VideoInfo, error) {
	result, err := c.FetchChannelVideosWithToken(ctx, channelID, "", maxResults)
	if err != nil {
		return nil, err
	}
	return result.Videos, nil
}

func (c *Client) FetchChannelVideosWithToken(ctx context.Context, channelID string, pageToken string, maxResults int64) (*FetchResult, error) {
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
	return c.FetchPlaylistVideosWithToken(ctx, uploadsPlaylistID, pageToken, maxResults)
}

func (c *Client) FetchPlaylistVideos(ctx context.Context, playlistID string, maxResults int64) ([]VideoInfo, error) {
	result, err := c.FetchPlaylistVideosWithToken(ctx, playlistID, "", maxResults)
	if err != nil {
		return nil, err
	}
	return result.Videos, nil
}

func (c *Client) FetchPlaylistVideosWithToken(ctx context.Context, playlistID string, pageToken string, maxResults int64) (*FetchResult, error) {
	call := c.service.PlaylistItems.List([]string{"snippet", "contentDetails"}).
		PlaylistId(playlistID).
		MaxResults(maxResults)

	if pageToken != "" {
		call = call.PageToken(pageToken)
	}

	resp, err := call.Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("youtube: fetch playlist: %w", err)
	}

	videoIDs := make([]string, 0, len(resp.Items))
	for _, item := range resp.Items {
		videoIDs = append(videoIDs, item.ContentDetails.VideoId)
	}

	if len(videoIDs) == 0 {
		return &FetchResult{Videos: nil, NextPageToken: ""}, nil
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
		videos = append(videos, VideoInfo{
			ID:           v.Id,
			Title:        v.Snippet.Title,
			ThumbnailURL: getBestThumbnail(v.Snippet.Thumbnails),
			Duration:     v.ContentDetails.Duration,
			PublishedAt:  publishedAt,
		})
	}
	return &FetchResult{Videos: videos, NextPageToken: resp.NextPageToken}, nil
}

func getBestThumbnail(t *youtube.ThumbnailDetails) string {
	if t == nil {
		return ""
	}
	if t.Medium != nil {
		return t.Medium.Url
	}
	if t.High != nil {
		return t.High.Url
	}
	if t.Default != nil {
		return t.Default.Url
	}
	return ""
}
