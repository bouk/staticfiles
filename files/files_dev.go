//+build dev

package files

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"os"
	"path"
	"time"
)

var staticDir = "static"

var (
	NotFound  = http.NotFound
	Server    = http.FileServer(http.Dir(staticDir))
	ServeHTTP = Server.ServeHTTP
)

func Open(name string) (io.ReadCloser, error) {
	return os.Open(filepath.Join(staticDir, name))
}

func ModTime(name string) (t time.Time) {
	stat, err := os.Stat(filepath.Join(staticDir, name))
	if err != nil {
		t = stat.ModTime()
	}
	return
}

func Hash(name string) (s string) {
	f, err := os.Open(filepath.Join(staticDir, name))
	if err != nil {
		return
	}
	defer f.Close()
	hash := sha256.New()
	io.Copy(hash, f)
	return hex.EncodeToString(hash.Sum(nil))
}

func OpenGlob(name string) ([]io.ReadCloser, error) {
	readers := make([]io.ReadCloser, 0)
	for file := range staticFiles {
		matches, err := path.Match(name, file)
		if err != nil {
			continue
		}
		if matches {
			reader, err := Open(file)
			if err == nil && reader != nil {
				readers = append(readers, reader)
			}
		}
	}
	if len(readers) == 0 {
		return nil, fmt.Errorf("No assets found that match.")
	}
	return readers, nil
}
