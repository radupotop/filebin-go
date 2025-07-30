package tests

import (
	"testing"

	"github.com/radupotop/filebin-go/pkg/backends"
)

func TestCheckFileShouldError(t *testing.T) {
	content := []byte("Hello world")
	fileName := "hello.txt"

	fileHeader := MockMultipartFileHeader(t, fileName, content)
	file, _ := fileHeader.Open()
	mimeType := backends.GetContentType(file)
	defer file.Close()

	// must error
	resp, err := backends.CheckFile(fileHeader, mimeType)
	if err == nil ||
		resp.IsError() == false ||
		resp.Message != "File type not allowed" {
		t.Error("Test Failed")
	}
}

func TestCheckFileAllowed(t *testing.T) {
	pngsig := []byte("\x89PNG\x0D\x0A\x1A\x0A")
	fileName := "demo.png"
	fileHeader := MockMultipartFileHeader(t, fileName, pngsig)
	file, _ := fileHeader.Open()
	mimeType := backends.GetContentType(file)
	defer file.Close()

	// must not error
	resp, err := backends.CheckFile(fileHeader, mimeType)
	if err != nil || resp.IsError() {
		t.Error("Check failed")
	}
}
