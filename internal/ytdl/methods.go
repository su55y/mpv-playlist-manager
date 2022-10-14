package ytdl

import (
	"errors"
	"log"
	"os/exec"
	"strings"
)

func GetVideoData(url string) (Video, error) {
	out, err := execute(url)
	if err != nil {
		return Video{}, err
	}
	return parseVideo(out)
}

func parseVideo(output string) (Video, error) {
	var v Video
	if s := strings.Split(output, "\n"); s != nil && len(s) > 1 {
		return Video{Title: s[0], Thumb: s[1]}, nil
	}
	return v, errors.New("can't parse output")
}

func execute(url string) (string, error) {
	args := []string{
		ytDlIgnoreErrors,
		ytDlGetTitle,
		ytDlGetThumbArg,
		ytDlNoWarnings,
		url,
	}
	out, err := exec.Command(ytDl, args...).Output()
	if err != nil {
		log.Println(err.Error())
		return "", err
	}

	return string(out), nil
}
