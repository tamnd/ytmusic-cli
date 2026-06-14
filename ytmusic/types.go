package ytmusic

// Song is a YouTube Music song/track.
type Song struct {
	Rank     int    `json:"rank"      csv:"rank"      tsv:"rank"`
	Title    string `json:"title"     csv:"title"     tsv:"title"`
	Artist   string `json:"artist"    csv:"artist"    tsv:"artist"`
	Album    string `json:"album"     csv:"album"     tsv:"album"`
	Duration string `json:"duration"  csv:"duration"  tsv:"duration"`
	VideoID  string `json:"video_id"  csv:"video_id"  tsv:"video_id"`
	URL      string `json:"url"       csv:"url"       tsv:"url"`
}

// Artist is a YouTube Music artist.
type Artist struct {
	Rank        int    `json:"rank"         csv:"rank"         tsv:"rank"`
	Name        string `json:"name"         csv:"name"         tsv:"name"`
	BrowseID    string `json:"browse_id"    csv:"browse_id"    tsv:"browse_id"`
	Subscribers string `json:"subscribers"  csv:"subscribers"  tsv:"subscribers"`
	URL         string `json:"url"          csv:"url"          tsv:"url"`
}

// Album is a YouTube Music album.
type Album struct {
	Rank     int    `json:"rank"      csv:"rank"      tsv:"rank"`
	Title    string `json:"title"     csv:"title"     tsv:"title"`
	Artist   string `json:"artist"    csv:"artist"    tsv:"artist"`
	Year     string `json:"year"      csv:"year"      tsv:"year"`
	BrowseID string `json:"browse_id" csv:"browse_id" tsv:"browse_id"`
	URL      string `json:"url"       csv:"url"       tsv:"url"`
}

// Playlist is a YouTube Music playlist.
type Playlist struct {
	Rank       int    `json:"rank"        csv:"rank"        tsv:"rank"`
	Title      string `json:"title"       csv:"title"       tsv:"title"`
	Author     string `json:"author"      csv:"author"      tsv:"author"`
	TrackCount string `json:"track_count" csv:"track_count" tsv:"track_count"`
	BrowseID   string `json:"browse_id"   csv:"browse_id"   tsv:"browse_id"`
	URL        string `json:"url"         csv:"url"         tsv:"url"`
}
