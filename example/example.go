package main

import (
	"bou.ke/staticfiles/files"
	"net/http"
)

func main() {
	http.ListenAndServe(":7684", files.Server)
}
