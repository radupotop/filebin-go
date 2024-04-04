package views

import (
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/radupotop/filebin-go/backends"
	"github.com/radupotop/filebin-go/marshal"
)

// Handler for rendering the HTML form
func FormHandler(w http.ResponseWriter, r *http.Request) {
	formHTML, _ := backends.ReadFile("templates/template.html")
	tmpl := template.Must(template.New("form").Parse(formHTML))
	tmpl.Execute(w, nil)
}

// Handler for uploading file to S3
func UploadHandler(w http.ResponseWriter, r *http.Request) {
	// Parse the multipart form
	err := r.ParseMultipartForm(backends.MAX_FILE_SIZE * 5)
	if err != nil {
		fmt.Println(err)
		resp := marshal.Response{Message: "Unable to parse form", Status: http.StatusBadRequest}
		resp.ReturnJson(w)
		return
	}

	// Get the map of uploaded files
	files := r.MultipartForm.File["files"]
	use_s3 := r.FormValue("s3") == "on"

	// Will store list of results
	var results marshal.ResponseResults

	// Iterate over each uploaded file
	for _, handler := range files {
		// Open the uploaded file
		file, err := handler.Open()

		if err != nil {
			fmt.Println(err)
			resp := marshal.Response{Message: "Failed to retrieve file", Status: http.StatusBadRequest}
			resp.ReturnJson(w)
			return
		}
		defer file.Close()

		resp, err := backends.CheckFile(handler)
		if err != nil {
			fmt.Println(err)
			resp.ReturnJson(w)
			return
		}

		var destFile string
		// Continue to S3 upload
		if use_s3 {
			destFile, err = backends.PutToS3(w, file, handler.Filename)
		} else {
			destFile, err = backends.CopyFileTemp(w, file)
		}
		if err != nil {
			fmt.Println(err)
			return
		}

		results = append(results, marshal.UpResult{Orig: handler.Filename, Dest: destFile})
	}

	log.Printf("Results: %+v", results)

	resp := marshal.Response{
		Message: "Files saved",
		Context: marshal.ResponseContext{"Use S3", fmt.Sprint(use_s3)},
		Results: results,
		Status:  http.StatusCreated,
	}
	resp.ReturnJson(w)
}
