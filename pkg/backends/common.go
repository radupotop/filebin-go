package backends

import (
	"io"
	"mime/multipart"
	"net/http"
)

// Get mime-type by detection of magic numbers.
// Only read the first 512 bytes to detect mime-type.
func GetContentType(file multipart.File) string {
	buf := make([]byte, 512)
	file.Read(buf)
	file.Seek(0, io.SeekStart)
	mimeType := http.DetectContentType(buf)
	return mimeType
}
