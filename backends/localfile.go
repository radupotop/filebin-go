package backends

import (
	"fmt"
	"io"
	"log"
	"math"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"slices"

	"github.com/radupotop/filebin-go/marshal"
)

/*
MAGIC NUMBERS    https://en.m.wikipedia.org/wiki/List_of_file_signatures

PNG := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
JPEG := []byte{0xff, 0xd8, 0xff}
GIF87a := []byte{0x47, 0x49, 0x46, 0x38, 0x37, 0x61}
GIF89a := []byte{0x47, 0x49, 0x46, 0x38, 0x39, 0x61}

http.DetectContentType(PNG)
*/

const (
	// 320 KiB max file size
	MAX_FILE_SIZE = 10 << 15
)

var (
	ALLOWED_MIME_TYPES = []string{"image/jpeg", "image/png", "image/gif"}
	FILE_SIZE_UNIT     = math.Pow(1024, 2) // MiB
	FS_UNIT_NAME       = "MiB"
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

	// Check file mime type
	mimeType := handler.Header.Get("Content-Type")
	if !slices.Contains(ALLOWED_MIME_TYPES, mimeType) {
		resp := marshal.Response{
			Message: "File type not allowed",
			Context: marshal.ResponseContext{
				handler.Filename,
				"Must be one of",
				fmt.Sprint(ALLOWED_MIME_TYPES),
				"Instead detected",
				mimeType,
			},
			Status: http.StatusUnsupportedMediaType,
		}
		return resp, fmt.Errorf("file type not allowed: %s", mimeType)
	}

	return marshal.RESP_OK, nil
}

// Check if the Mime type is the same across HTTP header, file extension and actual file content
func CheckMimeMatches(handler *multipart.FileHeader, file multipart.File) (marshal.Response, error) {

	// Only read the first 512 bytes to detect mime type.
	buf := make([]byte, 512)
	file.Read(buf)

	// Mime type by HTTP header
	mimeByHeader := handler.Header.Get("Content-Type")
	// Mime type by file extension
	mimeByExt := mime.TypeByExtension(filepath.Ext(handler.Filename))
	// Mime type by detection of magic numbers
	mimeByMagic := http.DetectContentType(buf)

	if mimeByHeader == mimeByExt && mimeByExt == mimeByMagic {
		return marshal.RESP_OK, nil
	} else {
		resp := marshal.Response{
			Message: "Mime type mismatch",
			Context: marshal.ResponseContext{
				"All must match",
				mimeByHeader,
				mimeByExt,
				mimeByMagic,
			},
			Status: http.StatusUnsupportedMediaType,
		}
		return resp, fmt.Errorf("mime type mismatch")
	}
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
