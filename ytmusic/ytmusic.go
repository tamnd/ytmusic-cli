// Package ytmusic provides access to YouTube Music via the InnerTube API.
package ytmusic

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Config holds client configuration.
type Config struct {
	BaseURL   string
	Rate      time.Duration
	Timeout   time.Duration
	Retries   int
	UserAgent string
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() Config {
	return Config{
		BaseURL:   "https://music.youtube.com",
		Rate:      300 * time.Millisecond,
		Timeout:   30 * time.Second,
		Retries:   3,
		UserAgent: "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	}
}

// Client fetches data from YouTube Music.
type Client struct {
	cfg  Config
	http *http.Client
	last time.Time
}

// NewClient creates a new Client.
func NewClient(cfg Config) *Client {
	return &Client{
		cfg:  cfg,
		http: &http.Client{Timeout: cfg.Timeout},
	}
}

// searchParams are the InnerTube filter params for each content type.
// These are base64-encoded protobuf values from the YTMusic web app.
const (
	paramsSongs     = "EgWKAQIIAWoKEAMQBBAKEAkQBQ=="
	paramsArtists   = "EgWKAQIgAWoKEAMQBBAKEAkQBQ=="
	paramsAlbums    = "EgWKAQIYAWoKEAMQBBAKEAkQBQ=="
	paramsPlaylists = "EgeKAQQoAEABagoQAxAEEAoQCRAF"
)

func (c *Client) post(ctx context.Context, path string, body map[string]any) ([]byte, error) {
	if c.cfg.Rate > 0 {
		if wait := c.cfg.Rate - time.Since(c.last); wait > 0 {
			select {
			case <-time.After(wait):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	var respBody []byte
	for attempt := 0; attempt <= c.cfg.Retries; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.cfg.BaseURL+path, bytes.NewReader(payload))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("User-Agent", c.cfg.UserAgent)
		req.Header.Set("Origin", c.cfg.BaseURL)
		req.Header.Set("Referer", c.cfg.BaseURL+"/")

		resp, err := c.http.Do(req)
		c.last = time.Now()
		if err != nil {
			if attempt < c.cfg.Retries {
				time.Sleep(time.Duration(attempt+1) * time.Second)
				continue
			}
			return nil, err
		}
		respBody, err = io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if err != nil {
			return nil, err
		}
		if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
			if attempt < c.cfg.Retries {
				time.Sleep(time.Duration(attempt+1) * time.Second)
				continue
			}
			return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
		}
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
		}
		break
	}
	return respBody, nil
}

// findAll recursively finds all values for a given key in a JSON-decoded structure.
func findAll(obj any, key string) []any {
	var results []any
	switch v := obj.(type) {
	case map[string]any:
		if val, ok := v[key]; ok {
			results = append(results, val)
		}
		for _, child := range v {
			results = append(results, findAll(child, key)...)
		}
	case []any:
		for _, item := range v {
			results = append(results, findAll(item, key)...)
		}
	}
	return results
}

func getStr(m map[string]any, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func flexColumnTexts(item map[string]any, colIndex int) []string {
	cols, _ := item["flexColumns"].([]any)
	if colIndex >= len(cols) {
		return nil
	}
	col, _ := cols[colIndex].(map[string]any)
	cr, _ := col["musicResponsiveListItemFlexColumnRenderer"].(map[string]any)
	text, _ := cr["text"].(map[string]any)
	runs, _ := text["runs"].([]any)
	var out []string
	for _, r := range runs {
		rm, _ := r.(map[string]any)
		if t := getStr(rm, "text"); t != "" && t != " • " {
			out = append(out, t)
		}
	}
	return out
}

func parseItems(data []byte) []map[string]any {
	var raw any
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil
	}
	items := findAll(raw, "musicResponsiveListItemRenderer")
	out := make([]map[string]any, 0, len(items))
	for _, item := range items {
		if m, ok := item.(map[string]any); ok {
			out = append(out, m)
		}
	}
	return out
}

func getVideoID(item map[string]any) string {
	for _, o := range findAll(item, "watchEndpoint") {
		if m, ok := o.(map[string]any); ok {
			if id := getStr(m, "videoId"); id != "" {
				return id
			}
		}
	}
	return ""
}

func getBrowseID(item map[string]any) string {
	for _, e := range findAll(item, "browseEndpoint") {
		if m, ok := e.(map[string]any); ok {
			if id := getStr(m, "browseId"); id != "" {
				return id
			}
		}
	}
	return ""
}

func buildSearchBody(query, params string) map[string]any {
	body := map[string]any{
		"context": map[string]any{
			"client": map[string]any{
				"clientName":    "WEB_REMIX",
				"clientVersion": "1.20231101",
			},
		},
		"query": query,
	}
	if params != "" {
		body["params"] = params
	}
	return body
}

// SearchSongs searches YouTube Music for songs.
func (c *Client) SearchSongs(ctx context.Context, query string, limit int) ([]Song, error) {
	if limit <= 0 {
		limit = 20
	}
	data, err := c.post(ctx, "/youtubei/v1/search?prettyPrint=false", buildSearchBody(query, paramsSongs))
	if err != nil {
		return nil, fmt.Errorf("search songs: %w", err)
	}
	items := parseItems(data)
	var out []Song
	for i, item := range items {
		if len(out) >= limit {
			break
		}
		col0 := flexColumnTexts(item, 0)
		col1 := flexColumnTexts(item, 1)
		title := ""
		if len(col0) > 0 {
			title = col0[0]
		}
		rest := col1
		if len(rest) > 0 && strings.EqualFold(rest[0], "song") {
			rest = rest[1:]
		}
		artist, album, duration := "", "", ""
		if len(rest) > 0 {
			artist = rest[0]
		}
		if len(rest) > 1 {
			album = rest[1]
		}
		if len(rest) > 2 {
			duration = rest[len(rest)-1]
		}
		vid := getVideoID(item)
		url := ""
		if vid != "" {
			url = "https://music.youtube.com/watch?v=" + vid
		}
		out = append(out, Song{
			Rank:     i + 1,
			Title:    title,
			Artist:   artist,
			Album:    album,
			Duration: duration,
			VideoID:  vid,
			URL:      url,
		})
	}
	return out, nil
}

// SearchArtists searches YouTube Music for artists.
func (c *Client) SearchArtists(ctx context.Context, query string, limit int) ([]Artist, error) {
	if limit <= 0 {
		limit = 20
	}
	data, err := c.post(ctx, "/youtubei/v1/search?prettyPrint=false", buildSearchBody(query, paramsArtists))
	if err != nil {
		return nil, fmt.Errorf("search artists: %w", err)
	}
	items := parseItems(data)
	var out []Artist
	for i, item := range items {
		if len(out) >= limit {
			break
		}
		col0 := flexColumnTexts(item, 0)
		col1 := flexColumnTexts(item, 1)
		name := ""
		if len(col0) > 0 {
			name = col0[0]
		}
		subs := ""
		if len(col1) > 1 {
			subs = col1[len(col1)-1]
		}
		bid := getBrowseID(item)
		url := ""
		if bid != "" {
			url = "https://music.youtube.com/channel/" + bid
		}
		out = append(out, Artist{
			Rank:        i + 1,
			Name:        name,
			BrowseID:    bid,
			Subscribers: subs,
			URL:         url,
		})
	}
	return out, nil
}

// SearchAlbums searches YouTube Music for albums.
func (c *Client) SearchAlbums(ctx context.Context, query string, limit int) ([]Album, error) {
	if limit <= 0 {
		limit = 20
	}
	data, err := c.post(ctx, "/youtubei/v1/search?prettyPrint=false", buildSearchBody(query, paramsAlbums))
	if err != nil {
		return nil, fmt.Errorf("search albums: %w", err)
	}
	items := parseItems(data)
	var out []Album
	for i, item := range items {
		if len(out) >= limit {
			break
		}
		col0 := flexColumnTexts(item, 0)
		col1 := flexColumnTexts(item, 1)
		title := ""
		if len(col0) > 0 {
			title = col0[0]
		}
		rest := col1
		if len(rest) > 0 && (strings.EqualFold(rest[0], "album") || strings.EqualFold(rest[0], "ep") || strings.EqualFold(rest[0], "single")) {
			rest = rest[1:]
		}
		year, artist := "", ""
		if len(rest) > 0 {
			year = rest[0]
		}
		if len(rest) > 1 {
			artist = rest[1]
		}
		bid := getBrowseID(item)
		url := ""
		if bid != "" {
			url = "https://music.youtube.com/browse/" + bid
		}
		out = append(out, Album{
			Rank:     i + 1,
			Title:    title,
			Artist:   artist,
			Year:     year,
			BrowseID: bid,
			URL:      url,
		})
	}
	return out, nil
}

// SearchPlaylists searches YouTube Music for playlists.
func (c *Client) SearchPlaylists(ctx context.Context, query string, limit int) ([]Playlist, error) {
	if limit <= 0 {
		limit = 20
	}
	data, err := c.post(ctx, "/youtubei/v1/search?prettyPrint=false", buildSearchBody(query, paramsPlaylists))
	if err != nil {
		return nil, fmt.Errorf("search playlists: %w", err)
	}
	items := parseItems(data)
	var out []Playlist
	for i, item := range items {
		if len(out) >= limit {
			break
		}
		col0 := flexColumnTexts(item, 0)
		col1 := flexColumnTexts(item, 1)
		title := ""
		if len(col0) > 0 {
			title = col0[0]
		}
		rest := col1
		if len(rest) > 0 && strings.EqualFold(rest[0], "playlist") {
			rest = rest[1:]
		}
		author, trackCount := "", ""
		if len(rest) > 0 {
			author = rest[0]
		}
		if len(rest) > 1 {
			trackCount = rest[len(rest)-1]
		}
		bid := getBrowseID(item)
		url := ""
		if bid != "" {
			url = "https://music.youtube.com/browse/" + bid
		}
		out = append(out, Playlist{
			Rank:       i + 1,
			Title:      title,
			Author:     author,
			TrackCount: trackCount,
			BrowseID:   bid,
			URL:        url,
		})
	}
	return out, nil
}
