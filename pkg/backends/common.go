package backends

import (
	"io"
	"math"
	"mime/multipart"
	"net/http"
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
	// 2.5 MiB max file size
	MAX_FILE_SIZE = 10 << 18
)

var (
	ALLOWED_MIME_TYPES = []string{"image/jpeg", "image/png", "image/gif"}
	FILE_SIZE_UNIT     = math.Pow(1024, 2) // MiB
	FS_UNIT_NAME       = "MiB"
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
