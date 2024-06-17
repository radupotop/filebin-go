package main

import (
	"log"
	"net/http"

	"github.com/radupotop/filebin-go/pkg/views"
)

func main() {
	http.HandleFunc("/", views.FormHandler)
	http.HandleFunc("/upload", views.UploadHandler)
	log.Println("Server started on port 8080")
	http.ListenAndServe(":8080", nil)
}
