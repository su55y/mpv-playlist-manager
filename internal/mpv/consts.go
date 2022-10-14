package mpv

const (
	DEFAULT_CMD = "mpv --idle --input-ipc-server=/tmp/mpv.sock"
	// commands
	APPEND_PLAY         = "append-play"
	APPEND              = "append"
	LOADFILE            = "loadfile"
	GET_PROPERTY        = "get_property"
	PLAYLIST            = "playlist"
	PLAYLIST_NEXT       = "playlist-next"
	PLAYLIST_PREV       = "playlist-prev"
	PLAY                = "play"
	PAUSE               = "pause"
	CYCLE               = "cycle"
	PLAYLIST_PLAY_INDEX = "playlist-play-index"

	MPV_SOCK = "mpv.sock"
)
