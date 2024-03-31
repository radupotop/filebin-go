package main

import (
	"fmt"
	"log"
	"math"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"slices"
)

const (
	// 320 KiB max file size
	MAX_FILE_SIZE = 10 << 15
)

var (
	ALLOWED_EXTENSIONS = []string{".png", ".jpg", ".jpeg"}
	FILE_SIZE_UNIT     = math.Pow(1024, 2) // MiB
)

// Read file from disk
func readFile(filename string) (string, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		log.Println("Error reading file:", err)
		return "", err
	}
	log.Printf("Loaded file: %s\n", filename)
	return string(content), nil
}

// Check if upload pre-conditions are met: such as file size and extension
func checkFile(w http.ResponseWriter, handler *multipart.FileHeader) {
	// Check file size
	if handler.Size > MAX_FILE_SIZE {
		resp := Response{
			Message: "File size exceeds the limit",
			Context: fmt.Sprintf("Max file size must be: %.2f MiB", MAX_FILE_SIZE/FILE_SIZE_UNIT),
			Status:  http.StatusRequestEntityTooLarge,
		}
		resp.returnJson(w)
		return
	}

	// Check file extension
	extension := filepath.Ext(handler.Filename)
	if !slices.Contains(ALLOWED_EXTENSIONS, extension) {
		resp := Response{
			Message: "File extension not allowed",
			Context: fmt.Sprintf("Must be one of %s", ALLOWED_EXTENSIONS),
			Status:  http.StatusUnsupportedMediaType,
		}
		resp.returnJson(w)
		return
	}
}
