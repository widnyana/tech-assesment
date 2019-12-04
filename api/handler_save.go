package api

import (
	"kumparan/internal/contract"
	"kumparan/internal/queue"
	"log"
	"net/http"
	"time"
)

func handleNewsEndpoint(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		expectJSON(http.HandlerFunc(createContent)).ServeHTTP(w, r)
	default:
		handleNewsGet(w, r)
	}

}

func createContent(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1048576)

	var incoming contract.NewsData

	jderr := jsonDecoder(r.Body, &incoming)
	if jderr.isError {
		log.Printf("error decoding json body: %s\n", jderr.message)
		http.Error(w, jderr.message, jderr.httpCode)
		return
	}
	incoming.Created = time.Now()

	msg, err := incoming.Marshal()
	if err != nil {
		log.Printf("cannot marshal NewsData to json")
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	// send to message queue
	if err := queue.Publish(msg); err != nil {
		// response json
		responseAsJSON(w, contract.HTTPResponse{
			Error:   true,
			Message: err.Error(),
			Meta:    nil,
		}, http.StatusInternalServerError)
		return
	}

	// response json
	responseAsJSON(w, contract.HTTPResponse{
		Error:   false,
		Message: "success",
		Meta:    incoming,
	}, http.StatusOK)
}

