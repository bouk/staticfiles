//+build dev

package files

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

var staticDir = "static"

var (
	Server    = http.FileServer(http.Dir(staticDir))
	ServeHTTP = Server.ServeHTTP
)

func Open(name string) (io.Reader, error) {
	return os.Open(filepath.Join(staticDir, name))
}

func ModTime(name string) (t time.Time) {
	stat, err := os.Stat(filepath.Join(staticDir, name))
	if err != nil {
		t = stat.ModTime()
	}
	return
}
