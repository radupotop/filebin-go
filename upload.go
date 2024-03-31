package main

import (
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

	// Will store list of results
	var results []UpResult

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

		respCp, tmpFileName, errCp := copyFileTemp(file)
		if errCp != nil {
			respCp.returnJson(w)
			return
		}
		results = append(results, UpResult{Orig: handler.Filename, Dest: tmpFileName})
	}

	// // Continue to S3 upload
	// use_s3 := r.FormValue("s3")
	// if use_s3 == "on" {
	// 	// Status up to this point
	// 	statusMsg := fmt.Sprintf("File %s saved locally as: %s", handler.Filename, tempFile.Name())
	// 	putToS3(w, file, handler, statusMsg)
	// 	return
	// }

	log.Println("Results", results)

	resp := Response{
		Message: "Files saved locally",
		Context: results,
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
