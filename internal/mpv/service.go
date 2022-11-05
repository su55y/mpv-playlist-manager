package mpv

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"

	"github.com/su55y/yt2mpv/internal/ytdl"
)

type Service struct {
	path    string
	notify  bool
	Decoder *json.Decoder
	Logger  *log.Logger
}

var (
	currentPlaylist []PlaylistItem
	playlistCommand = &PlaylistCommand{Command: [2]string{GET_PROPERTY, PLAYLIST}}
	appendCommand   = func(file string) *AppendCommand {
		return &AppendCommand{Command: [3]string{LOADFILE, file, APPEND_PLAY}, Async: true}
	}
	playIndexCommand = func(index string) *PlayIndexCommand {
		return &PlayIndexCommand{Command: [2]string{PLAYLIST_PLAY_INDEX, index}}
	}
	nextCommand       = &NextPrevCommand{Command: [1]string{PLAYLIST_NEXT}}
	prevCommand       = &NextPrevCommand{Command: [1]string{PLAYLIST_PREV}}
	pauseCycleCommand = &PauseCycleCommand{Command: [2]string{CYCLE, PAUSE}}
)

func NewService(notify bool) Service {
	logger := log.New(os.Stdout, "mpv: ", log.LstdFlags)
	path := filepath.Join(os.TempDir(), MPV_SOCK)
	if _, err := os.Stat(path); err != nil && errors.Is(err, os.ErrNotExist) {
		logger.Fatalf("socket file not exists: %s", err.Error())
	}
	return Service{
		Logger: logger,
		path:   path,
		notify: notify,
	}
}

func (s *Service) PlayIndexCtl(index string) error {
	conn, err := s.getConn()
	if err != nil {
		return err
	}
	go func() {
		var resp DefaultResponse
		if !decodeData(&resp, conn) {
			msg := "can't read play index response"
			s.Logger.Println(msg)
			return
		}
		s.Logger.Printf("play index response: %#+v", resp)
	}()

	sendCommand(playIndexCommand(index), conn)
	time.Sleep(1e9)
	return err
}

func (s *Service) PlayIndex(index string, w *http.ResponseWriter) error {
	conn, err := s.getConn()
	if err != nil {
		return err
	}

	i, err := strconv.Atoi(index)
	if err != nil {
		return err
	}

	if len(currentPlaylist) <= i {
		return fmt.Errorf("invalid index: %d", i)
	}

	go func() {
		var resp DefaultResponse
		if !decodeData(&resp, conn) {
			s.Logger.Printf("err: %s\n", err.Error())
			return
		}
		s.Logger.Printf("play index response: %#+v", resp)
		err = json.NewEncoder(*w).Encode(&resp)
	}()

	sendCommand(playIndexCommand(index), conn)
	time.Sleep(1e9)
	return err
}

func (s *Service) DoAction(action string, w *http.ResponseWriter) error {
	conn, err := s.getConn()
	if err != nil {
		return err
	}

	go func() {
		var resp DefaultResponse
		if !decodeData(&resp, conn) {
			s.Logger.Printf("err: %s\n", err.Error())
			return
		}
		s.Logger.Printf("action response: %#+v", resp)
		err = json.NewEncoder(*w).Encode(&resp)
	}()

	switch action {
	case "pause", "play":
		sendCommand(pauseCycleCommand, conn)
	case "next":
		sendCommand(nextCommand, conn)
	case "prev":
		sendCommand(prevCommand, conn)
	default:
		return fmt.Errorf("unknown error")
	}

	time.Sleep(1e9)
	return err
}

func (s *Service) AppendVideo(url string, w *http.ResponseWriter) error {
	conn, err := s.getConn()
	if err != nil {
		return err
	}

	go func() {
		var resp DefaultResponse
		if !decodeData(&resp, conn) {
			s.Logger.Printf("err: %s\n", err.Error())
			return
		}
		// s.Logger.Printf("append response: %#+v", resp)
		if resp.Err == "success" && s.notify {
			s.sendNotify(fmt.Sprintf("vid #%d just added", resp.Data.PlEntryId))
		}
		err = json.NewEncoder(*w).Encode(&resp)
		go updatePlaylistItem(url)
	}()
	sendCommand(appendCommand(url), conn)

	time.Sleep(1e9)
	return err
}

func (s *Service) GetPlaylistString() string {
	if err := s.requestPlaylist(); err != nil {
		s.Logger.Printf("[Service.GetPlaylist]playlist request error: %s", err.Error())
		return ""
	}
	response := ""
	for i, v := range currentPlaylist {
		response += fmt.Sprintf("%s\000info\037%d\n", v.Title, i)
	}
	return response
}

func (s *Service) GetPlaylist(w *http.ResponseWriter) {
	if err := s.requestPlaylist(); err != nil {
		s.Logger.Printf("[Service.GetPlaylist]playlist request error: %s", err.Error())
		return
	}
	err := json.NewEncoder(*w).Encode(&currentPlaylist)
	if err != nil {
		s.Logger.Printf("can't encode playlist: %s", err.Error())
	}
}

func (s *Service) GetPlaylistLength() int {
	return len(currentPlaylist)
}

func (s *Service) requestPlaylist() error {
	conn, err := s.getConn()
	if err != nil {
		return err
	}

	go func() {
		var resp Playlist
		if !decodeData(&resp, conn) {
			s.Logger.Printf("err: %s\n", err.Error())
			return
		}
		// s.Logger.Printf("playlist response: %v", resp)
		go whichCurrent(resp)
	}()
	sendCommand(playlistCommand, conn)

	time.Sleep(1e9)

	return err
}

func (s *Service) getConn() (net.Conn, error) {
	conn, err := net.Dial("unix", s.path)
	if err != nil {
		s.Logger.Printf("can't connect to socket: %s", err.Error())
		return nil, err
	}
	return conn, nil
}

func decodeData[T MpvRequests](data *T, conn net.Conn) bool {
	err := json.NewDecoder(conn).Decode(data)
	return err == nil
}

func sendCommand[T MpvCommands](c *T, conn net.Conn) error {
	raw, err := json.Marshal(c)
	if err != nil {
		return err
	}
	raw = append(raw, byte('\n'))
	if _, err := conn.Write(raw); err != nil {
		return err
	}
	return nil
}

func updatePlaylistItem(url string) {
	plItem := PlaylistItem{
		Filename: url,
	}
	v, err := ytdl.GetVideoData(url)
	if err == nil {
		plItem.Title = v.Title
		plItem.Thumbnail = v.Thumb
	}
	currentPlaylist = append(currentPlaylist, plItem)
}

func whichCurrent(resp Playlist) {
	if resp.Data != nil {
		for i, v := range resp.Data {
			if len(currentPlaylist) > i && v.Current {
				for j := range currentPlaylist {
					currentPlaylist[j].Current = false
				}
				currentPlaylist[i].Current = true
			}
		}
	}
}

func getIndex(item PlaylistItem) int {
	for i, v := range currentPlaylist {
		if item.Filename == v.Filename {
			return i
		}
	}
	return -1
}

func (s *Service) sendNotify(msg string) {
	if err := exec.Command(
		"notify-send",
		[]string{"-i", "mpv", "-a", "mpv", msg}...,
	).Start(); err != nil {
		s.Logger.Printf("notify-send failed: %s", err.Error())
	}
}
