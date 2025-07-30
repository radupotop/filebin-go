package views

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/radupotop/filebin-go/pkg/backends"
	"github.com/radupotop/filebin-go/pkg/marshal"
)

const (
	UPLOAD_TEMPLATE_FILE string = "templates/template.html"
)

// Handler for rendering the HTML form
func FormHandler(w http.ResponseWriter, r *http.Request) {
	formHTML, _ := backends.ReadFile(UPLOAD_TEMPLATE_FILE)
	tmpl := template.Must(template.New("form").Parse(formHTML))
	tmpl.Execute(w, nil)
}

// Handler for uploading file to S3
func UploadHandler(w http.ResponseWriter, r *http.Request) {
	begin := time.Now()
	var waitgroup sync.WaitGroup
	var resp marshal.Response
	// Parse the multipart form
	err := r.ParseMultipartForm(backends.MAX_FILE_SIZE * 5)
	if err != nil {
		log.Println(err)
		resp := marshal.Response{Message: "Unable to parse form", Status: http.StatusBadRequest}
		resp.ReturnJson(w)
		return
	}

	// Get the map of uploaded files
	files := r.MultipartForm.File["files"]
	use_s3 := r.FormValue("s3") == "on"
	num_files := len(files)

	// Will store list of results
	var results marshal.ResponseResults
	// Use a buffered channel
	errChan := make(chan marshal.Response, num_files)
	log.Printf("Received %d files\n", num_files)

	// Iterate over each uploaded file
	for idx, header := range files {
		// Open the uploaded file
		file, err := header.Open()

		if err != nil {
			log.Println(err)
			resp = marshal.Response{Message: "Failed to retrieve file", Status: http.StatusBadRequest}
			resp.ReturnJson(w)
			return
		}
		defer file.Close()

		mimeType := backends.GetContentType(file)
		resp, err = backends.CheckFile(header, mimeType)
		if err != nil {
			log.Println(err)
			resp.ReturnJson(w)
			return
		}

		var destFile string
		// Continue to S3 upload
		if use_s3 {
			destKey := backends.GenUuidFilename(header.Filename)
			destFile = backends.GenPublicFileURL(destKey)
			waitgroup.Add(1)
			go backends.PutToS3(errChan, file, destKey, mimeType, &waitgroup, idx+1)
		} else {
			destFile, err = backends.CopyFileTemp(w, file)
			if err != nil {
				log.Println(err)
				return
			}
		}
		upres := marshal.UpResult{Orig: header.Filename, Dest: destFile, MimeType: mimeType}
		results = append(results, upres)
	}
	waitgroup.Wait()
	// Channel must be closed, or the range loop will block
	close(errChan)

	if len(errChan) > 0 {
		log.Printf("ErrChan queue: %d", len(errChan))
	}

	// Only the first error will be returned via the API
	for resp := range errChan {
		if resp.IsError() {
			resp.ReturnJson(w)
			return
		}
	}

	log.Printf("Results: %+v", results)

	resp = marshal.Response{
		Message: "Files saved",
		Context: marshal.ResponseContext{"Use S3", fmt.Sprint(use_s3)},
		Results: results,
		Status:  http.StatusCreated,
	}
	resp.ReturnJson(w)
	log.Printf("All uploads finished in: %s\n", time.Since(begin))
}
