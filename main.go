package main

import (
	"context"
	"crypto/sha1"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"regexp"
	"strconv"
	"strings"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/su55y/yt2mpv/internal/mpv"
)

type key int

const (
	requestIDKey key = 0
)

var (
	healthy    int32
	port       string
	notify     bool
	service    mpv.Service
	rxURL      = regexp.MustCompile("^https?.+\\.\\w+\\/.+$")
	runMpvArgs []string
)

func init() {
	flag.StringVar(&port, "p", "5000", "port")
	flag.BoolVar(&notify, "n", false, "send notifications")

	var runMpvCmd string
	flag.StringVar(&runMpvCmd, "m", mpv.DEFAULT_CMD, "mpv command")
	flag.Parse()
	if runMpvArgs = strings.Split(runMpvCmd, " "); len(runMpvArgs) < 2 || runMpvArgs[0] != "mpv" {
		log.Fatal(fmt.Errorf("invalid mpv command: (%#+v)", runMpvCmd))
	}
}

func main() {
	if err := syscall.Mknod("/tmp/mpv.sock", syscall.S_IFSOCK|0666, 0); err != nil {
		log.Fatal(err)
	}
	logger := log.New(os.Stdout, "main: ", log.LstdFlags)
	logger.Printf("start %q", runMpvArgs)
	c1 := exec.Command("mpv", runMpvArgs[1:]...)
	if err := c1.Start(); err != nil {
		log.Fatal()
	}
	time.Sleep(3 * time.Second)

	service = mpv.NewService(notify)
	logger.Println("service is starting...")

	logger.Println("server is starting...")
	router := http.NewServeMux()
	router.Handle("/", index())
	router.Handle("/req", appendVid())
	router.Handle("/playlist", getPlaylist())
	router.Handle("/play", playIndex())
	router.Handle("/control", control())
	router.Handle("/healthz", healthz())
	router.Handle("/rofi", getPlaylistRofi())

	// nextRequestID := func() string {
	// 	return fmt.Sprintf("%d", time.Now().Unix())
	// }
	nextRequestID := func() string {
		b := make([]byte, 16)
		copy(b[:], fmt.Sprintf("%d", time.Now().UnixMicro()))

		return fmt.Sprintf("%x", sha1.Sum(b))
	}

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      tracing(nextRequestID)(logging(logger)(router)),
		ErrorLog:     logger,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	done := make(chan bool)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)

	go func() {
		<-quit
		logger.Println("server is shutting down...")
		atomic.StoreInt32(&healthy, 0)

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		server.SetKeepAlivesEnabled(false)
		if err := server.Shutdown(ctx); err != nil {
			logger.Fatalf("could not gracefully shutdown the server: %v\n", err)
		}
		close(done)
	}()
	go func() {
		if err := c1.Wait(); err != nil {
			logger.Fatal(err)
		}
		logger.Println("mpv process successfully exited")
		os.Exit(0)
	}()

	logger.Printf("server is ready to handle requests at :%s", port)
	atomic.StoreInt32(&healthy, 1)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Fatalf("could not listen on %s: %v\n", port, err)
	}

	<-done
	logger.Println("server stopped")
}

func index() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	})
}

func appendVid() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		urlParam, ok := r.URL.Query()["u"]
		if !ok || urlParam == nil || len(urlParam) != 1 || !rxURL.MatchString(urlParam[0]) {
			json.NewEncoder(w).Encode(&mpv.ErrorResponse{Err: true, Message: "invalid url"})
		} else {
			service.AppendVideo(urlParam[0], &w)
		}
	})
}

func getPlaylist() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		service.GetPlaylist(&w)
	})
}

func getPlaylistRofi() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.UserAgent() != "rofi" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		if index, ok := r.URL.Query()["index"]; ok {
			if i, err := strconv.Atoi(index[0]); err == nil &&
				service.GetPlaylistLength() > i &&
				i >= 0 {
				if err := service.PlayIndexCtl(index[0]); err != nil {
					w.WriteHeader(http.StatusBadRequest)
				}
				return
			}
			w.WriteHeader(400)
			w.Write([]byte("bad index"))
			return
		}

		playlist := service.GetPlaylistString()
		w.WriteHeader(200)
		if _, err := w.Write([]byte(playlist)); err != nil {
			service.Logger.Printf("[Service.GetPlaylistRofi]write response error: %s", err.Error())
		}
	})
}

func control() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if action, ok := r.URL.Query()["action"]; ok && len(action) == 1 {
			switch action[0] {
			case "pause", "play", "next", "prev":
				service.DoAction(action[0], &w)
			}
		} else {
			fmt.Println("unknown action")
			w.WriteHeader(http.StatusBadRequest)

		}
	})
}

func playIndex() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if i, ok := r.URL.Query()["index"]; ok && len(i) == 1 {
			if _, err := strconv.Atoi(i[0]); err == nil {
				if err := service.PlayIndex(i[0], &w); err != nil {
					json.NewEncoder(w).Encode(&mpv.ErrorResponse{Err: true, Message: err.Error()})
				}
				return
			}
		}
		json.NewEncoder(w).Encode(&mpv.ErrorResponse{Err: true, Message: "invalid index"})
	})
}

func healthz() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch atomic.LoadInt32(&healthy) {
		case 1:
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusServiceUnavailable)
		}
	})
}

func logging(logger *log.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				requestID, ok := r.Context().Value(requestIDKey).(string)
				if !ok {
					logger.Printf("unknown request id: %s", requestID)
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				logger.Printf("%s %s %s", r.Method, r.URL.String(), requestID)
			}()
			next.ServeHTTP(w, r)
		})
	}
}

func tracing(nextRequestID func() string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}
			requestID := r.Header.Get("X-Request-Id")
			if requestID == "" {
				requestID = nextRequestID()
			}
			ctx := context.WithValue(r.Context(), requestIDKey, requestID)
			w.Header().Set("X-Request-Id", requestID)
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Content-Type", "application/json")
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
