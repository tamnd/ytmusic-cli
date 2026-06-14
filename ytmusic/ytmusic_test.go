package ytmusic_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/tamnd/ytmusic-cli/ytmusic"
)

func newTestClient(ts *httptest.Server) *ytmusic.Client {
	cfg := ytmusic.DefaultConfig()
	cfg.BaseURL = ts.URL
	cfg.Rate = 0
	return ytmusic.NewClient(cfg)
}

// buildSearchResponse builds a fake InnerTube search response.
func buildSearchResponse(items []map[string]any) []byte {
	data, _ := json.Marshal(map[string]any{
		"contents": map[string]any{
			"tabbedSearchResultsRenderer": map[string]any{
				"tabs": []any{
					map[string]any{
						"tabRenderer": map[string]any{
							"content": map[string]any{
								"sectionListRenderer": map[string]any{
									"contents": []any{
										map[string]any{
											"musicShelfRenderer": map[string]any{
												"contents": items,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	})
	return data
}

func makeItem(title, artist, album, duration, videoID string) map[string]any {
	return map[string]any{
		"musicResponsiveListItemRenderer": map[string]any{
			"flexColumns": []any{
				map[string]any{
					"musicResponsiveListItemFlexColumnRenderer": map[string]any{
						"text": map[string]any{
							"runs": []any{
								map[string]any{"text": title},
							},
						},
					},
				},
				map[string]any{
					"musicResponsiveListItemFlexColumnRenderer": map[string]any{
						"text": map[string]any{
							"runs": []any{
								map[string]any{"text": "Song"},
								map[string]any{"text": " • "},
								map[string]any{"text": artist},
								map[string]any{"text": " • "},
								map[string]any{"text": album},
								map[string]any{"text": " • "},
								map[string]any{"text": duration},
							},
						},
					},
				},
			},
			"overlay": map[string]any{
				"musicItemThumbnailOverlayRenderer": map[string]any{
					"content": map[string]any{
						"musicPlayButtonRenderer": map[string]any{
							"playNavigationEndpoint": map[string]any{
								"watchEndpoint": map[string]any{
									"videoId": videoID,
								},
							},
						},
					},
				},
			},
		},
	}
}

func TestSearchSongs(t *testing.T) {
	items := []map[string]any{
		makeItem("Jazz Song 1", "Artist A", "Album X", "3:45", "abc123"),
		makeItem("Jazz Song 2", "Artist B", "Album Y", "4:12", "def456"),
	}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/search") {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, string(buildSearchResponse(items)))
	}))
	defer ts.Close()

	c := newTestClient(ts)
	songs, err := c.SearchSongs(context.Background(), "jazz", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(songs) != 2 {
		t.Fatalf("got %d songs, want 2", len(songs))
	}
	if songs[0].Title != "Jazz Song 1" {
		t.Errorf("title = %q, want 'Jazz Song 1'", songs[0].Title)
	}
	if songs[0].Artist != "Artist A" {
		t.Errorf("artist = %q, want 'Artist A'", songs[0].Artist)
	}
	if songs[0].VideoID != "abc123" {
		t.Errorf("videoID = %q, want 'abc123'", songs[0].VideoID)
	}
}
