package marshal

import (
	"encoding/json"
	"net/http"
)

type ResponseContext []string
type ResponseResults []UpResult

type Response struct {
	Message string          `json:"message"`
	Context ResponseContext `json:"context"`
	Results ResponseResults `json:"results"`
	Status  int             `json:"status"`
}

// Return JSON response
func (r *Response) ReturnJson(w http.ResponseWriter) {
	// Set the Content-Type header and the HTTP status code
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(r.Status)
	// Marshall struct to JSON
	json.NewEncoder(w).Encode(r)
}

// Individual upload result
type UpResult struct {
	Orig string `json:"orig"`
	Dest string `json:"dest"`
}

// // All results
// type UploadResults struct {
// 	Results []UpResult `json:"results"`
// }

// func (u *UploadResults) add(upres UpResult) {
// 	u.Results = append(u.Results, upres)
// }

// func (u *UploadResults) toJson() (string, error) {
// 	resp, err := json.Marshal(u)
// 	return string(resp), err
// }

// // MarshalJSON implements the json.Marshaler interface for Response
// func (r Response) MarshalJSON() ([]byte, error) {
// 	// Define a map to hold the message
// 	data := map[string]string{
// 		"message": r.Message,
// 		"status":  r.Status,
// 	}
// 	// Marshal the map into JSON
// 	return json.Marshal(data)
// }

// func ReturnJson(w http.ResponseWriter, resp Response) {
// 	json.NewEncoder(w).Encode(resp)
// }
