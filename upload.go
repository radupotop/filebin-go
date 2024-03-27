package main

import (
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"slices"
)

const (
	// 10 MB max file size
	MAX_FILE_SIZE = 10 << 20
)

var (
	ALLOWED_EXTENSIONS = []string{".png", ".jpg", ".jpeg"}
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
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return
	}

	// Retrieve the file from the form data
	file, handler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Failed to retrieve file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Check file size
	if handler.Size > MAX_FILE_SIZE {
		http.Error(w, "File size exceeds the limit", http.StatusBadRequest)
		return
	}

	// Check file extension
	extension := filepath.Ext(handler.Filename)
	if !slices.Contains(ALLOWED_EXTENSIONS, extension) {
		http.Error(w, fmt.Sprintf("File extension not allowed %s", extension), http.StatusBadRequest)
		return
	}

	// Create a new file in the server's temporary directory
	tempFile, err := os.CreateTemp("", "upload-*")
	if err != nil {
		http.Error(w, "Unable to create temporary file", http.StatusInternalServerError)
		return
	}
	defer tempFile.Close()

	// Copy the file content to the temporary file
	_, err = io.Copy(tempFile, file)
	if err != nil {
		http.Error(w, "Unable to copy file content", http.StatusInternalServerError)
		return
	}

	// Status up to this point
	statusMsg := fmt.Sprintf("File %s saved locally as: %s", handler.Filename, tempFile.Name())

	// Continue to S3 upload
	use_s3 := r.FormValue("s3")
	if use_s3 == "on" {
		putToS3(w, file, handler, statusMsg)
		return
	}

	fmt.Fprint(w, statusMsg)
}

func main() {
	http.HandleFunc("/", formHandler)
	http.HandleFunc("/upload", uploadHandler)
	fmt.Println("Server started on port 8080")
	http.ListenAndServe(":8080", nil)
}
