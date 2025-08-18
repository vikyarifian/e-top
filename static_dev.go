//+build dev
//go:build new
// +build new

package main

import (
	"fmt"
	"net/http"
	"os"
)

func Public() http.Handler {
	fmt.Println("building static files for development")
	return http.StripPrefix("/public", http.FileServerFs(os.DirFs("public")))
}
