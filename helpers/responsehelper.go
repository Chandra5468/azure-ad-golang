package helpers

import (
	"encoding/json"
	"net/http"
)

type ErrorFormat struct {
	Status int    `json:"status"`
	Data   string `json:"data"`
}

func ErrorFormatter(w http.ResponseWriter, statusCode int, err error) {
	// w.Write()
	// HeadersAdder(w)
	ef := ErrorFormat{
		Status: statusCode,
		Data:   err.Error(),
	}
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(&ef)

	// log.Print(w)
}

func ResponseFormatter(w http.ResponseWriter, statusCode int, data any) {
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(&data)
}

// Try to pass this in main.go file
// http.ResponseWriter is an interface. If you pass interface in a function under the hood it is passing pointer. So, no worries.
// func HeadersAdder(w http.ResponseWriter) {
// 	w.Header().Add("Content-Type", "application/json")
// 	w.Header().Add("Access-Control-Allow-Origin", "*")
// 	w.Header().Add("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, PATCH, DELETE")

// }
