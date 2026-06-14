# ytmusic

Search [YouTube Music](https://music.youtube.com) songs, artists, albums, and playlists from the command line.

`ytmusic` is a single pure-Go binary. No API key required.

## Install

```bash
go install github.com/tamnd/ytmusic-cli/cmd/ytmusic@latest
```

Or grab a prebuilt binary from the [releases](https://github.com/tamnd/ytmusic-cli/releases), or run the container image:

```bash
docker run --rm ghcr.io/tamnd/ytmusic:latest --help
```

## Usage

```bash
# Search for songs (default)
ytmusic search "jazz"
ytmusic search "lo-fi hip hop"

# Search for artists
ytmusic search "skrillex" --kind artists

# Search for albums
ytmusic search "dark side of the moon" --kind albums

# Search for playlists
ytmusic search "workout" --kind playlists

# Output formats
ytmusic search "jazz" -o json
ytmusic search "classical" -o csv -n 10
ytmusic search "rock" -o table
```

## Commands

| Command | Description |
|---------|-------------|
| `search <query>` | Search YouTube Music for songs, artists, albums, or playlists |
| `version` | Show version information |

## Search kinds

`songs` (default), `artists`, `albums`, `playlists`

## Global flags

```
-o, --output string    output format: table|json|jsonl|csv|tsv|url|raw (default "auto")
-n, --limit int        limit number of records (0 = command default: 20)
    --fields strings   comma-separated columns to include
    --no-header        omit header row
    --template string  Go text/template per record
    --timeout duration per-request timeout (default 30s)
    --delay duration   minimum spacing between requests
```

## License

Apache-2.0. See [LICENSE](LICENSE).
