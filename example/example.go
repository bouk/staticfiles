package main

import (
	"github.com/bouk/staticfiles/files"
	"net/http"
)

func main() {
	http.ListenAndServe(":7684", files.Server)
}
