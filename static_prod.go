//go:build !dev
// +build !dev

package main

import (
	"embed"
	"net/http"
)

//go:embed public
var publicFs embed.FS

func public() http.Handler {
	return http.FileServerFS(publicFs)
}
