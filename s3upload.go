package main

import (
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/google/uuid"
	"github.com/subosito/gotenv"
)

const (
	uploadACL string = "public-read"
)

var (
	awsRegion   string
	awsS3Bucket string
)

// Upload to S3 bucket
func putToS3(w http.ResponseWriter, multipartFile multipart.File, origFilename string) string {
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(awsRegion),
		Credentials: credentials.NewEnvCredentials(),
	})

	if err != nil {
		resp := Response{
			Message: "Failed to create AWS session",
			Status:  http.StatusInternalServerError,
		}
		resp.returnJson(w)
		return ""
	}

	svc := s3.New(sess)
	uuidFilename := "img/" + uuid.New().String() + filepath.Ext(origFilename)

	// Upload file to S3 bucket
	_, err = svc.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(awsS3Bucket),
		Key:    aws.String(uuidFilename),
		Body:   multipartFile,
		// ACL:    aws.String(uploadACL),
	})
	if err != nil {
		resp := Response{
			Message: "Failed to upload file to S3 bucket",
			Context: ResponseContext{uuidFilename, awsS3Bucket},
			Status:  http.StatusInternalServerError,
		}
		resp.returnJson(w)
		return ""
	}
	return uuidFilename
}

// only called once
func init() {
	gotenv.Must(gotenv.Load)
	log.Println("Loading env vars")
	awsRegion = os.Getenv("AWS_REGION")
	awsS3Bucket = os.Getenv("AWS_BUCKET")
	// The following are read by NewEnvCredentials
	awsAccessKeyID := os.Getenv("AWS_ACCESS_KEY_ID")
	// "AWS_SECRET_ACCESS_KEY"
	log.Println("Using env vars:", [3]string{awsRegion, awsS3Bucket, awsAccessKeyID})
}
