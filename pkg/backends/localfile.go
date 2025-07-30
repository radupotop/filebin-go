package backends

import (
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"slices"

	"github.com/radupotop/filebin-go/pkg/marshal"
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
func CheckFile(header *multipart.FileHeader, mimeType string) (marshal.Response, error) {
	// Check file size
	if header.Size > MAX_FILE_SIZE {
		resp := marshal.Response{
			Message: "File size exceeds the limit",
			Context: marshal.ResponseContext{
				header.Filename,
				"Max file size must be",
				fmt.Sprintf("%.2f %s", MAX_FILE_SIZE/FILE_SIZE_UNIT, FS_UNIT_NAME),
			},
			Status: http.StatusRequestEntityTooLarge,
		}
		return resp, fmt.Errorf("file size %d exceeds the limit: %d bytes", header.Size, MAX_FILE_SIZE)
	}

	if !slices.Contains(ALLOWED_MIME_TYPES, mimeType) {
		resp := marshal.Response{
			Message: "File type not allowed",
			Context: marshal.ResponseContext{
				header.Filename,
				"Must be one of",
				fmt.Sprint(ALLOWED_MIME_TYPES),
				"Instead detected",
				mimeType,
			},
			Status: http.StatusUnsupportedMediaType,
		}
		return resp, fmt.Errorf("file type not allowed: %s", mimeType)
	}

	// All good
	return marshal.RESP_OK, nil
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
