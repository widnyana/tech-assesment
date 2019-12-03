package api

import (
	"kumparan/internal/contract"
	"log"
	"net/http"
	"strings"
)

func expectJSON(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctype := r.Header.Get("Content-Type")
		if strings.ToLower(ctype) != "application/json" {
			log.Printf("unexpected content type: %s\n", ctype)
			responseAsJSON(w, contract.HTTPResponse{
				Message: "do you speak json?",
				Error:   true,
			}, http.StatusBadRequest)
			return
		}
		next.ServeHTTP(w, r)
	})
}
