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

const INTERVAL = 500 * time.Millisecond

// Custom writer to record progress
type PassThru struct {
	Reader       io.Reader
	Completed    int64
	InstallFlags cstructs.InstallFlags
	DownloadURL  string
	Last         int64
	Total        int64
	SpeedsN      int
	SpeedSum     float32
	LastChecked  time.Time
}

// This function gets called by io.Copy
func (pt *PassThru) Read(p []byte) (int, error) {
	n, err := pt.Reader.Read(p)
	pt.Completed += int64(n)

	if time.Now().Sub(pt.LastChecked) >= INTERVAL {
		percent := float64(pt.Completed) / float64(pt.Total) * 100
		speed := float32(pt.Completed-pt.Last) / float32(INTERVAL.Seconds())
		pt.SpeedsN++
		pt.SpeedSum += speed
		pt.Last = pt.Completed

		avgSpeed := int64(pt.SpeedSum) / int64(pt.SpeedsN)

		var eta int64
		if avgSpeed > 0 {
			eta = (pt.Total - pt.Completed) / avgSpeed
		}

		hours := eta / 3600
		minutes := (eta % 3600) / 60
		seconds := eta % 60

		if err == nil {
			if pt.InstallFlags.OutputStyle == "default" {
				fmt.Printf("\r\033[2KDownloading game files ... \t%.2f%% (%d / %d bytes) @ %.2f MiB/s ETA %d:%02d:%02d", percent, pt.Completed, pt.Total, float64(speed)/1024/1024, hours, minutes, seconds)
			} else if pt.InstallFlags.OutputStyle == "json" {
				fmt.Printf("\r\033[2K{\"task\": \"Downloading game files\", \"details\": \"Downloading %v\", \"progress\": %v, \"eta\": %v}", pt.DownloadURL, percent, eta)
			}
		}

		pt.LastChecked = time.Now()
	}

	return n, err
}

// Downloads the file from 'downloadURL' into the game's install directory
func Download(downloadURL string, installFlags cstructs.InstallFlags, gameMeta cstructs.GameMeta) error {
	r, err := http.Get(downloadURL)
	if err != nil {
		return fmt.Errorf("Download: Unable to make request to download URL: '%v': %v", downloadURL, err)
	}
	defer r.Body.Close()

	u, err := url.Parse(downloadURL)
	if err != nil {
		return fmt.Errorf("Download: Unable to parse URL: '%v'", downloadURL)
	}

	filename := path.Base(u.Path)

	file, err := os.Create(filepath.Join(gameMeta.GetDirectory(), filename))
	if err != nil {
		return fmt.Errorf("Download: Unable to create file: '%v'", filepath.Join(gameMeta.GetDirectory(), filename))
	}
	defer file.Close()

	src := &PassThru{Reader: r.Body, InstallFlags: installFlags, DownloadURL: downloadURL, Total: r.ContentLength}

	_, err = io.Copy(file, src)
	if err != nil {
		return fmt.Errorf("Download: Unable to write to file: '%v'", filepath.Join(gameMeta.GetDirectory(), filename))
	}

	return nil
}

// Downloads the file(s) from 'downloadURL' into the game's install directory through BitTorrent
func DownloadTorrentMagnet(downloadURL string, logger logging.Logger, installFlags cstructs.InstallFlags, gameMeta cstructs.GameMeta) error {
	cfg := torrent.NewDefaultClientConfig()
	cfg.DataDir = gameMeta.GetDirectory()

	// Replace anacrolix's wanky logger for a simpler one
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
		return fmt.Errorf("DownloadTorrentMagnet: Unable to create client")
	}

	t, err := client.AddMagnet(downloadURL)
	if err != nil {
		closeErrs := client.Close()
		if closeErrs != nil {
			return fmt.Errorf("DownloadTorrentMagnet: Error while closing client: %v", errors.Join(closeErrs...))
		}

		return fmt.Errorf("DownloadTorrentMagnet: Unable to add magnet: '%v'", downloadURL)
	}

	if installFlags.NoSeeding {
		t.DisallowDataUpload()
	}

	<-t.GotInfo()

	// Allow for clean interruption (although it usually works very poorly)
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c

		t.Drop()
		client.Close()

		os.Exit(1)
	}()

	t.DownloadAll()

	ticker := time.NewTicker(INTERVAL)
	defer ticker.Stop()
	last := t.BytesCompleted()

	speedsN := 0
	var speedSum float32 = 0
	for range ticker.C {
		info := t.Info()

		total := info.TotalLength()
		completed := t.BytesCompleted()
		percent := float64(completed) / float64(total) * 100
		speed := float32(completed-last) / float32(INTERVAL.Seconds())
		speedsN++
		speedSum += speed
		last = completed

		avgSpeed := int64(speedSum) / int64(speedsN)

		var eta int64
		if avgSpeed > 0 {
			eta = (total - completed) / avgSpeed
		}

		hours := eta / 3600
		minutes := (eta % 3600) / 60
		seconds := eta % 60

		if installFlags.OutputStyle == "default" {
			fmt.Printf("\r\033[2KDownloading game files ... \t%.2f%% (%d / %d bytes) @ %.2f MiB/s ETA %d:%02d:%02d", percent, completed, total, float64(speed)/1024/1024, hours, minutes, seconds)
		} else if installFlags.OutputStyle == "json" {
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
		return fmt.Errorf("DownloadTorrentMagnet: Error while closing client: %v", errors.Join(errs...))
	}

	// anacrolix/torrent leaves a SQLite file behind with the torrent's info
	torrentFile, err := filepath.Glob(filepath.Join(gameMeta.GetDirectory(), ".torrent*"))
	if err != nil {
		return nil
	}

	// Remove the torrent info file
	err = os.Remove(torrentFile[0])
	if err != nil {
		return fmt.Errorf("DownloadTorrentMagnet: Unable to delete torrent file: %v", err)
	}

	return nil
}
