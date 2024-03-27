package main

import (
	"fmt"
	"html/template"
	"net/http"

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
	file, handler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Failed to retrieve file from form", http.StatusBadRequest)
		return
	}
	defer file.Close()

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
