package backends

import (
	"fmt"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/google/uuid"
	"github.com/subosito/gotenv"

	"github.com/radupotop/filebin-go/marshal"
)

var (
	awsRegion   string
	awsS3Bucket string
	uploadACL   string
	uploadDir   string // upload dir inside the bucket
)

// Generate a new filename based on UUID4 + the original file extension
func GenUuidFilename(origFilename string) string {
	return uploadDir + uuid.New().String() + filepath.Ext(origFilename)
}

// Upload to S3 bucket
func PutToS3(respChan chan marshal.Response, multipartFile multipart.File, destFilename string, waitgroup *sync.WaitGroup, idx int) {

	defer waitgroup.Done()

	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(awsRegion),
		Credentials: credentials.NewEnvCredentials(),
	})

	if err != nil {
		resp := marshal.Response{
			Message: "Failed to create AWS session",
			Status:  http.StatusInternalServerError,
			Err:     fmt.Errorf("failed to create AWS session"),
		}
		respChan <- resp
		return
	}

	svc := s3.New(sess)

	// Upload file to S3 bucket
	_, err = svc.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(awsS3Bucket),
		Key:    aws.String(destFilename),
		Body:   multipartFile,
		ACL:    aws.String(uploadACL),
	})
	if err != nil {
		resp := marshal.Response{
			Message: "Failed to upload file to S3 bucket",
			Context: marshal.ResponseContext{destFilename, awsS3Bucket, uploadACL},
			Status:  http.StatusInternalServerError,
			Err:     fmt.Errorf("failed to upload file to S3 bucket"),
		}
		respChan <- resp
		return
	}
	respChan <- marshal.RESP_OK
	log.Printf("S3 upload #%d finished: %s", idx+1, destFilename)
}

// only called once
func init() {
	gotenv.Must(gotenv.Load)
	log.Println("Loading env vars")
	awsRegion = os.Getenv("AWS_REGION")
	awsS3Bucket = os.Getenv("AWS_BUCKET")
	uploadACL = os.Getenv("UPLOAD_ACL")
	uploadDir = os.Getenv("UPLOAD_DIR")
	// The following are read by NewEnvCredentials
	// They are only checked here.
	_, keyIsSet := os.LookupEnv("AWS_ACCESS_KEY_ID")
	_, secretIsSet := os.LookupEnv("AWS_SECRET_ACCESS_KEY")
	log.Println("Using env vars:", []string{
		awsRegion, awsS3Bucket, uploadACL, uploadDir,
		fmt.Sprint(keyIsSet && secretIsSet),
	})
}
