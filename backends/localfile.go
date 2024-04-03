package backends

import (
	"fmt"
	"io"
	"log"
	"math"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"slices"

	"github.com/radupotop/filebin-go/marshal"
)

const (
	// 320 KiB max file size
	MAX_FILE_SIZE = 10 << 15
)

var (
	ALLOWED_EXTENSIONS = []string{".png", ".jpg", ".jpeg"}
	FILE_SIZE_UNIT     = math.Pow(1024, 2) // MiB
	FS_UNIT_NAME       = "MiB"
	RESP_OK            = marshal.Response{Message: "OK", Status: http.StatusOK}
)

// Read file from disk
func ReadFile(filename string) (string, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		log.Println("Error reading file:", err)
		return "", err
	}
	log.Println("Loaded file:", filename)
	return string(content), nil
}

// Check if pre-conditions are met for upload
func CheckFile(handler *multipart.FileHeader) (marshal.Response, error) {
	// Check file size
	if handler.Size > MAX_FILE_SIZE {
		resp := marshal.Response{
			Message: "File size exceeds the limit",
			Context: marshal.ResponseContext{
				handler.Filename,
				"Max file size must be",
				fmt.Sprintf("%.2f %s", MAX_FILE_SIZE/FILE_SIZE_UNIT, FS_UNIT_NAME),
			},
			Status: http.StatusRequestEntityTooLarge,
		}
		return resp, fmt.Errorf("file size %d exceeds the limit: %d bytes", handler.Size, MAX_FILE_SIZE)
	}

	// Check file extension
	extension := filepath.Ext(handler.Filename)
	if !slices.Contains(ALLOWED_EXTENSIONS, extension) {
		resp := marshal.Response{
			Message: "File extension not allowed",
			Context: marshal.ResponseContext{
				handler.Filename,
				"Must be one of",
				fmt.Sprint(ALLOWED_EXTENSIONS),
			},
			Status: http.StatusUnsupportedMediaType,
		}
		return resp, fmt.Errorf("file extension not allowed: %s", extension)
	}

	return RESP_OK, nil
}

func CopyFileTemp(w http.ResponseWriter, file multipart.File) (string, error) {
	// Create a new file in the server's temporary directory
	tempFile, err := os.CreateTemp("", "upload-*")
	if err != nil {
		resp := marshal.Response{
			Message: "Unable to create temporary file",
			Status:  http.StatusInternalServerError,
		}
		resp.ReturnJson(w)
		return "", err
	}
	tempFileName := tempFile.Name()
	defer tempFile.Close()

	// Copy the file content to the temporary file
	_, err = io.Copy(tempFile, file)
	if err != nil {
		resp := marshal.Response{
			Message: "Unable to copy file content",
			Context: marshal.ResponseContext{tempFileName},
			Status:  http.StatusInternalServerError,
		}
		resp.ReturnJson(w)
		return "", err
	}

	return tempFileName, nil
}