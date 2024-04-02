package main

import (
	"fmt"
	"io"
	"io/fs"
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
	FS_UNIT_NAME       = "MiB"
	RESP_OK            = Response{Message: "OK", Status: http.StatusOK}
)

// Read file from disk
func readFile(filename string) (string, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		log.Println("Error reading file:", err)
		return "", err
	}
	log.Println("Loaded file:", filename)
	return string(content), nil
}

// Check if pre-conditions are met for upload
func checkFile(handler *multipart.FileHeader) (Response, error) {
	// Check file size
	if handler.Size > MAX_FILE_SIZE {
		resp := Response{
			Message: "File size exceeds the limit",
			Context: ResponseContext{
				handler.Filename,
				"Max file size must be",
				fmt.Sprintf("%.2f %s", MAX_FILE_SIZE/FILE_SIZE_UNIT, FS_UNIT_NAME),
			},
			Status: http.StatusRequestEntityTooLarge,
		}
		return resp, fs.ErrInvalid
	}

	// Check file extension
	extension := filepath.Ext(handler.Filename)
	if !slices.Contains(ALLOWED_EXTENSIONS, extension) {
		resp := Response{
			Message: "File extension not allowed",
			Context: ResponseContext{
				handler.Filename,
				"Must be one of",
				fmt.Sprint(ALLOWED_EXTENSIONS),
			},
			Status: http.StatusUnsupportedMediaType,
		}
		return resp, fs.ErrInvalid
	}

	return RESP_OK, nil
}

func copyFileTemp(file multipart.File) (Response, string, error) {
	// Create a new file in the server's temporary directory
	tempFile, err := os.CreateTemp("", "upload-*")
	if err != nil {
		resp := Response{Message: "Unable to create temporary file", Status: http.StatusInternalServerError}
		return resp, tempFile.Name(), err
	}
	defer tempFile.Close()

	// Copy the file content to the temporary file
	_, err = io.Copy(tempFile, file)
	if err != nil {
		resp := Response{Message: "Unable to copy file content", Status: http.StatusInternalServerError}
		return resp, tempFile.Name(), err
	}

	return RESP_OK, tempFile.Name(), nil
}
