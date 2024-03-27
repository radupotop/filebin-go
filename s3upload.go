package main

import (
	"fmt"
	"mime/multipart"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/subosito/gotenv"
)

var (
	awsRegion      string
	awsAccessKeyID string
	awsSecretKey   string
	s3Bucket       string
)

// Upload to S3 bucket
func putToS3(w http.ResponseWriter, file multipart.File, handler *multipart.FileHeader, prevMsg string) {
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
		http.Error(w, fmt.Sprintf("%s\nFailed to upload file to S3 bucket", prevMsg), http.StatusInternalServerError)
		return
	}

	fmt.Fprint(w, "File uploaded successfully!")
}

// only called once
func init() {
	gotenv.Load()
	fmt.Println("Loading env vars")
	awsRegion = os.Getenv("AWS_REGION")
	awsAccessKeyID = os.Getenv("AWS_KEY_ID")
	awsSecretKey = os.Getenv("AWS_SECRET")
	s3Bucket = os.Getenv("AWS_BUCKET")
}
