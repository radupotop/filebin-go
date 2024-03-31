package main

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
)

// Handler for rendering the HTML form
func formHandler(w http.ResponseWriter, r *http.Request) {
	formHTML, _ := readFile("template.html")
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
	// files := r.MultipartForm.File["file"]
	// fmt.Println(files)

	if err != nil {
		resp := Response{Message: "Failed to retrieve file", Status: http.StatusBadRequest}
		resp.returnJson(w)
		return
	}

	checkFile(w, handler)

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
	http.HandleFunc("/", formHandler)
	http.HandleFunc("/upload", uploadHandler)
	log.Println("Server started on port 8080")
	http.ListenAndServe(":8080", nil)
}
