package tests

import (
	"bytes"
	"mime"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"testing"
)

func MockMultipartFileHeader(t *testing.T, fileName string, content []byte) *multipart.FileHeader {
	// Mark the calling function as a test helper function.
	t.Helper()

	// Create a buffer and a multipart writer.
	var buffer bytes.Buffer
	writer := multipart.NewWriter(&buffer)

	// Create a form file with the given fileName.
	part, err := writer.CreateFormFile("file", fileName)
	if err != nil {
		t.Fatal(err)
	}

	// Write content to the part.
	part.Write(content)

	// Close the writer to finalize the boundary.
	writer.Close()

	/*
		// Alternative approach

		request := &http.Request{
			Header: http.Header{"Content-Type": []string{"multipart/form-data; boundary=" + writer.Boundary()}},
			Body:   io.NopCloser(&buffer),
		}
		request.ParseMultipartForm(32 << 20)
		fileHeader := request.MultipartForm.File["file"][0]
	*/

	// Create a multipart reader from the buffer.
	reader := multipart.NewReader(&buffer, writer.Boundary())

	// Read the next part (which should be our file).
	form, err := reader.ReadForm(32 << 20) // Max memory
	if err != nil {
		t.Fatal(err)
	}
	fileHeader := form.File["file"][0]

	// Add mime-type info
	mimeByExt := mime.TypeByExtension(filepath.Ext(fileName))
	mimeByDCT := http.DetectContentType(content)
	if mimeByExt != mimeByDCT {
		panic("Detected Content-Type different from extension")
	}
	fileHeader.Header.Set("Content-Type", mimeByDCT)
	// log.Println(fileHeader.Header)

	return fileHeader
}
