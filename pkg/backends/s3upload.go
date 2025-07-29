package backends

import (
	"fmt"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/google/uuid"
	"github.com/subosito/gotenv"

	"github.com/radupotop/filebin-go/pkg/marshal"
)

// If your bucket uses the bucket owner enforced setting for Object Ownership, ACLs are disabled and no longer affect permissions. All objects written to the bucket by any account will be owned by the bucket owner.
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

// Generate AWS URL to the file.
// https://BUCKET.s3.REGION.amazonaws.com/folder/file.ext
func GenFileURL(filename string) string {
	_url := &url.URL{
		Scheme: "https",
		Host:   awsS3Bucket + ".s3." + awsRegion + ".amazonaws.com",
		Path:   filename,
	}
	return _url.String()
}

// Upload to S3 bucket
func PutToS3(errChan chan marshal.Response, srcFile multipart.File, destKey string, mimeType string, waitgroup *sync.WaitGroup, idx int) {

	defer waitgroup.Done()

	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(awsRegion),
		Credentials: credentials.NewEnvCredentials(),
	})

	if err != nil {
		resp := marshal.Response{
			Message: "Failed to create AWS session",
			Status:  http.StatusInternalServerError,
		}
		log.Println("Failed to create AWS session")
		errChan <- resp
		return
	}

	svc := s3.New(sess)

	// Upload file to S3 bucket
	_, err = svc.PutObject(&s3.PutObjectInput{
		ACL:         aws.String(uploadACL),
		Body:        srcFile,
		Bucket:      aws.String(awsS3Bucket),
		ContentType: aws.String(mimeType),
		Key:         aws.String(destKey),
	})
	if err != nil {
		resp := marshal.Response{
			Message: "Failed to upload file to S3 bucket",
			Context: marshal.ResponseContext{destKey, awsS3Bucket, uploadACL},
			Status:  http.StatusInternalServerError,
		}
		log.Printf("_%d_ Failed to upload file to S3 bucket: %s -> %s\n", idx, destKey, awsS3Bucket)
		errChan <- resp
		return
	}
	log.Printf("_%d_ S3 upload finished: %s", idx, destKey)
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
