package mpv

type AppendCommand struct {
	Command [3]string `json:"command"`
	Async   bool      `json:"async,omitempty"`
}

type NextPrevCommand struct {
	Command [1]string `json:"command"`
}

type PlayIndexCommand struct {
	Command [2]string `json:"command"`
}

type PauseCycleCommand struct {
	Command [2]string `json:"command"`
}

type DefaultResponse struct {
	Data  RespData `json:"data"`
	Err   string   `json:"error"`
	ReqId int      `json:"request_id"`
}

type RespData struct {
	PlEntryId int `json:"playlist_entry_id"`
}

type PlaylistCommand struct {
	Command [2]string `json:"command"`
}

type Playlist struct {
	Data []PlaylistItem `json:"data"`
	Err  string         `json:"error"`
}

type PlaylistItem struct {
	Id        int    `json:"id"`
	Filename  string `json:"filename"`
	Current   bool   `json:"current"`
	Playing   bool   `json:"playing"`
	Title     string `json:"title"`
	Thumbnail string `json:"thumbnail"`
}

type ErrorResponse struct {
	Err     bool   `json:"error"`
	Message string `json:"message"`
}

type MpvCommands interface {
	AppendCommand | PlaylistCommand | NextPrevCommand | PauseCycleCommand | PlayIndexCommand
}

type MpvRequests interface {
	Playlist | DefaultResponse
}
