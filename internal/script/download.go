package script

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"syscall"
	"time"

	"codeberg.org/9g10f/marauderctl/internal/cstructs"
	"codeberg.org/9g10f/marauderctl/internal/logging"
	"github.com/anacrolix/log"
	"github.com/anacrolix/torrent"
)

func Download(downloadURL string, gameMeta cstructs.GameMeta) error {
	r, err := http.Get(downloadURL)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	u, _ := url.Parse(downloadURL)
	filename := path.Base(u.Path)

	file, err := os.Create(filepath.Join(gameMeta.GetDirectory(), filename))
	if err != nil {
		return err
	}

	_, err = io.Copy(file, r.Body)
	if err != nil {
		file.Close()
		return err
	}

	err = file.Close()
	if err != nil {
		return err
	}

	return nil
}

func DownloadTorrentMagnet(downloadURL string, outputStyle string, logger logging.Logger, gameMeta cstructs.GameMeta) error {
	cfg := torrent.NewDefaultClientConfig()
	cfg.DataDir = gameMeta.GetDirectory()
	cfg.Logger = log.NewLogger()
	cfg.Logger.SetHandlers(log.StreamHandler{
		W: logger,
		Fmt: func(r log.Record) []byte {
			if !r.Level.LessThan(log.Warning) {
				return []byte(r.Msg.String())
			} else {
				return []byte{}
			}
		},
	})

	client, err := torrent.NewClient(cfg)
	if err != nil {
		return err
	}

	t, err := client.AddMagnet(downloadURL)
	if err != nil {
		client.Close()
		return err
	}

	<-t.GotInfo()

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
        <-c

        t.Drop()
		client.Close()

		os.Exit(1)
    }()

	t.DownloadAll()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	var last int64 = t.BytesCompleted()

	speedsN := 0
	var speedSum int64 = 0
	for range ticker.C {
		info := t.Info()

		total := info.TotalLength()
		completed := t.BytesCompleted()
		percent := float64(completed) / float64(total) * 100
		speed := completed - last
		speedsN++
		speedSum += speed
		last = completed

		avgSpeed := speedSum / int64(speedsN)

		var eta int64 = 0
		if avgSpeed > 0 {
			eta = (total - completed) / avgSpeed
		}

		hours := eta / 3600
		minutes := (eta % 3600) / 60
		seconds := eta % 60

		if outputStyle == "default" {
			fmt.Printf("\r\033[2KDownloading game files ...   %.2f%% (%d / %d bytes) @ %.2f MiB/s ETA %d:%02d:%02d", percent, completed, total, float64(speed) / 1024 / 1024, hours, minutes, seconds)
		} else if outputStyle == "json" {
			fmt.Printf("\r\033[2K{\"task\": \"Downloading game files\", \"details\": \"Downloading %v\", \"progress\": %v, \"eta\": %v}", downloadURL, percent, eta)
		}

		if completed == total {
			break
		}
	}

	client.WaitAll()
	t.Drop()

	errs := client.Close()
	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	torrentFile, err := filepath.Glob(filepath.Join(gameMeta.GetDirectory(), ".torrent*"))
	if err != nil {
		return err
	}

	err = os.Remove(torrentFile[0])
	if err != nil {
		return err
	}

	return nil
}