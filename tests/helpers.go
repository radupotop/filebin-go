package tests

import (
	"bytes"
	"mime/multipart"
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

	// Close the writer to finalize the multipart message.
	writer.Close()

	// Create a multipart reader from the buffer.
	reader := multipart.NewReader(&buffer, writer.Boundary())

	// Read the next part (which should be our file).
	form, err := reader.ReadForm(32 << 20) // Max memory
	if err != nil {
		t.Fatal(err)
	}
	fileHeader := form.File["file"][0]

	return fileHeader
}
