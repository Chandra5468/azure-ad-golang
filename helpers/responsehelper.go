package helpers

import (
	"log"
	"net/http"
)

func HandlerErrors(w http.ResponseWriter, statusCode int, err error) {
	// w.Write()
	HeadersAdder(w)

	log.Print(w)
}

// http.ResponseWriter is an interface. If you pass interface in a function under the hood it is passing pointer. So, no worries.
func HeadersAdder(w http.ResponseWriter) {
	w.Header().Add("Content-Type", "application/json")
	w.Header().Add("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, PATCH, DELETE")

}
