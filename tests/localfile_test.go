package tests

import (
	"testing"

	"github.com/radupotop/filebin-go/backends"
)

func TestCheckFileShouldError(t *testing.T) {
	content := []byte("Hello world")
	fileName := "hello.txt"

	fileHeader := MockMultipartFileHeader(t, fileName, content)

	// must error
	resp, err := backends.CheckFile(fileHeader)
	if err == nil ||
		resp.IsError() == false ||
		resp.Message != "File extension not allowed" {
		t.Error("Test Failed")
	}
}

func TestCheckFileAllowed(t *testing.T) {
	content := []byte("Hello world")
	fileName := "hello.jpg"

	fileHeader := MockMultipartFileHeader(t, fileName, content)

	// must not error
	resp, err := backends.CheckFile(fileHeader)
	if err != nil || resp.IsError() {
		t.Error("Check failed")
	}

}
