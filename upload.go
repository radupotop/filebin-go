package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
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

	// Get the map of uploaded files
	files := r.MultipartForm.File["files"]
	use_s3 := r.FormValue("s3") == "on"

	// Will store list of results
	var results ResponseResults

	// Iterate over each uploaded file
	for _, handler := range files {
		// Open the uploaded file
		file, err := handler.Open()

		if err != nil {
			resp := Response{Message: "Failed to retrieve file", Status: http.StatusBadRequest}
			resp.returnJson(w)
			return
		}
		defer file.Close()

		respCheck, errCheck := checkFile(handler)
		if errCheck != nil {
			respCheck.returnJson(w)
			return
		}

		var destFile string
		// Continue to S3 upload
		if use_s3 {
			destFile, err = putToS3(w, file, handler.Filename)
		} else {
			destFile, err = copyFileTemp(w, file)
		}
		if err != nil {
			fmt.Println(err)
			return
		}

		results = append(results, UpResult{Orig: handler.Filename, Dest: destFile})
	}

	log.Println("Results:", results)

	resp := Response{
		Message: "Files saved",
		Context: ResponseContext{"Use S3", fmt.Sprint(use_s3)},
		Results: results,
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
