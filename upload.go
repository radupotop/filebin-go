package main

import (
	"fmt"
	"html/template"
	"io"
	"math"
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
)

// HTML form template
const formHTML = `
<!DOCTYPE html>
<html>
<head>
	<title>Upload File to S3</title>
</head>
<body>
	<form action="/upload" method="post" enctype="multipart/form-data">
		<input type="file" name="file">
		<input type="checkbox" id="s3" name="s3" checked>
		<label for="s3">Upload to S3</label>
		<input type="submit" value="Upload">
	</form>
</body>
</html>
`

// Handler for rendering the HTML form
func formHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.New("form").Parse(formHTML))
	tmpl.Execute(w, nil)
}

// Handler for uploading file to S3
func uploadHandler(w http.ResponseWriter, r *http.Request) {
	// Parse the multipart form
	err := r.ParseMultipartForm(MAX_FILE_SIZE)
	if err != nil {
		resp := Response{Message: "Unable to parse form", Status: http.StatusBadRequest}
		resp.returnJson(w)
		return
	}

	// Retrieve the file from the form data
	file, handler, err := r.FormFile("file")
	if err != nil {
		resp := Response{Message: "Failed to retrieve file", Status: http.StatusBadRequest}
		resp.returnJson(w)
		return
	}

	// Check file size
	if handler.Size > MAX_FILE_SIZE {
		resp := Response{
			Message: "File size exceeds the limit",
			Context: fmt.Sprintf("Max file size must be: %.2f MiB", MAX_FILE_SIZE/FILE_SIZE_UNIT),
			Status:  http.StatusBadRequest,
		}
		resp.returnJson(w)
		return
	}

	// Check file extension
	extension := filepath.Ext(handler.Filename)
	if !slices.Contains(ALLOWED_EXTENSIONS, extension) {
		resp := Response{
			Message: "File extension not allowed",
			Context: fmt.Sprintf("Must be one of %s", ALLOWED_EXTENSIONS),
			Status:  http.StatusBadRequest,
		}
		resp.returnJson(w)
		return
	}

	// Create a new file in the server's temporary directory
	tempFile, err := os.CreateTemp("", "upload-*")
	if err != nil {
		resp := Response{Message: "Unable to create temporary file", Status: http.StatusInternalServerError}
		resp.returnJson(w)
		return
	}

	defer file.Close()
	defer tempFile.Close()

	// Copy the file content to the temporary file
	_, err = io.Copy(tempFile, file)
	if err != nil {
		resp := Response{Message: "Unable to copy file content", Status: http.StatusInternalServerError}
		resp.returnJson(w)
		return
	}

	// Continue to S3 upload
	use_s3 := r.FormValue("s3")
	if use_s3 == "on" {
		// Status up to this point
		statusMsg := fmt.Sprintf("File %s saved locally as: %s", handler.Filename, tempFile.Name())
		putToS3(w, file, handler, statusMsg)
		return
	}

	resp := Response{
		Message: "File saved locally",
		Context: tempFile.Name(),
		Status:  http.StatusCreated,
	}
	resp.returnJson(w)
}

func main() {
	fmt.Printf("Using env vars: %s, %s\n", awsRegion, awsAccessKeyID)
	http.HandleFunc("/", formHandler)
	http.HandleFunc("/upload", uploadHandler)
	fmt.Println("Server started on port 8080")
	http.ListenAndServe(":8080", nil)
}
