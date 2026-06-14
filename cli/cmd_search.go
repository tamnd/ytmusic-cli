package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func (a *App) searchCmd() *cobra.Command {
	var kind string
	cmd := &cobra.Command{
		Use:   "search <query>",
		Short: "Search YouTube Music for songs, artists, albums, or playlists",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			query := args[0]
			limit := a.effectiveLimit(20)
			switch kind {
			case "songs", "song", "tracks", "track", "":
				songs, err := a.client.SearchSongs(cmd.Context(), query, limit)
				if err != nil {
					return mapFetchErr(err)
				}
				return a.renderOrEmpty(songs, len(songs))
			case "artists", "artist":
				artists, err := a.client.SearchArtists(cmd.Context(), query, limit)
				if err != nil {
					return mapFetchErr(err)
				}
				return a.renderOrEmpty(artists, len(artists))
			case "albums", "album":
				albums, err := a.client.SearchAlbums(cmd.Context(), query, limit)
				if err != nil {
					return mapFetchErr(err)
				}
				return a.renderOrEmpty(albums, len(albums))
			case "playlists", "playlist":
				playlists, err := a.client.SearchPlaylists(cmd.Context(), query, limit)
				if err != nil {
					return mapFetchErr(err)
				}
				return a.renderOrEmpty(playlists, len(playlists))
			default:
				return codeError(exitUsage, fmt.Errorf("unknown kind %q: use songs|artists|albums|playlists", kind))
			}
		},
	}
	cmd.Flags().StringVarP(&kind, "kind", "k", "songs", "what to search: songs|artists|albums|playlists")
	return cmd
}
