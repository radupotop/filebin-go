package main

import (
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

// Replace these with your AWS credentials and S3 bucket information
const (
	awsRegion      = "your-aws-region"
	awsAccessKeyID = "your-access-key-id"
	awsSecretKey   = "your-secret-key"
	s3Bucket       = "your-s3-bucket-name"
)

const (
	// 10 MB max file size
	MAX_FILE_SIZE = 10 << 20
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
	if extension != ".png" && extension != ".jpg" {
		http.Error(w, "File extension not allowed", http.StatusBadRequest)
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

	// Respond to the client indicating successful upload
	fmt.Fprintf(w, "File %s uploaded successfully as %s", handler.Filename, tempFile.Name())
	return

	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(awsRegion),
		Credentials: credentials.NewStaticCredentials(awsAccessKeyID, awsSecretKey, ""),
	})
	if err != nil {
		http.Error(w, "Failed to create AWS session", http.StatusInternalServerError)
		return
	}

	svc := s3.New(sess)

	// Upload file to S3 bucket
	_, err = svc.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(s3Bucket),
		Key:    aws.String(handler.Filename),
		Body:   file,
		ACL:    aws.String("public-read"),
	})
	if err != nil {
		http.Error(w, "Failed to upload file to S3", http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "File uploaded successfully!")
}

func main() {
	http.HandleFunc("/", formHandler)
	http.HandleFunc("/upload", uploadHandler)
	fmt.Println("Server started on port 8080")
	http.ListenAndServe(":8080", nil)
}
