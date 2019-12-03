package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

type (
	jsonDecodeResult struct {
		httpCode int
		message  string
		isError  bool
	}
)

// stolen with some modification from: https://www.alexedwards.net/blog/how-to-properly-parse-a-json-request-body
func jsonDecoder(source io.Reader, dest interface{}) jsonDecodeResult {
	dec := json.NewDecoder(source)
	dec.DisallowUnknownFields()

	err := dec.Decode(&dest)
	if err != nil {
		log.Println("jsonDecoder error")
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError

		switch {
		// Catch any syntax errors in the JSON and send an error message
		// which interpolates the location of the problem to make it
		// easier for the client to fix.
		case errors.As(err, &syntaxError):
			return jsonDecodeResult{
				isError:  true,
				httpCode: http.StatusBadRequest,
				message:  fmt.Sprintf("Request body contains badly-formed JSON (at position %d)", syntaxError.Offset),
			}

		// In some circumstances Decode() may also return an
		// io.ErrUnexpectedEOF error for syntax errors in the JSON. There
		// is an open issue regarding this at
		// https://github.com/golang/go/issues/25956.
		case errors.Is(err, io.ErrUnexpectedEOF):
			return jsonDecodeResult{
				isError:  true,
				httpCode: http.StatusBadRequest,
				message:  fmt.Sprintf("Request body contains badly-formed JSON"),
			}

		// Catch any type errors, like trying to assign a string in the
		// JSON request body to a int field in our Person struct. We can
		// interpolate the relevant field name and position into the error
		// message to make it easier for the client to fix.
		case errors.As(err, &unmarshalTypeError):
			return jsonDecodeResult{
				isError:  true,
				httpCode: http.StatusBadRequest,
				message: fmt.Sprintf("Request body contains an invalid value for the %q field (at position %d)",
					unmarshalTypeError.Field, unmarshalTypeError.Offset),
			}

		// Catch the error caused by extra unexpected fields in the request
		// body. We extract the field name from the error message and
		// interpolate it in our custom error message. There is an open
		// issue at https://github.com/golang/go/issues/29035 regarding
		// turning this into a sentinel error.
		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			return jsonDecodeResult{
				isError:  true,
				httpCode: http.StatusBadRequest,
				message:  fmt.Sprintf("Request body contains unknown field %s", fieldName),
			}

		// An io.EOF error is returned by Decode() if the request body is
		// empty.
		case errors.Is(err, io.EOF):
			return jsonDecodeResult{
				isError:  true,
				httpCode: http.StatusBadRequest,
				message:  "Request body must not be empty",
			}

		// Catch the error caused by the request body being too large. Again
		// there is an open issue regarding turning this into a sentinel
		// error at https://github.com/golang/go/issues/30715.
		case err.Error() == "http: request body too large":
			return jsonDecodeResult{
				isError:  true,
				httpCode: http.StatusRequestEntityTooLarge,
				message:  "Request body must not be larger than 1MB",
			}

		// Otherwise default to logging the error and sending a 500 Internal
		// Server Error response.
		default:
			log.Println(err.Error())
			return jsonDecodeResult{
				isError:  true,
				httpCode: http.StatusInternalServerError,
				message:  http.StatusText(http.StatusInternalServerError),
			}
		}
	}

	if dec.More() {
		return jsonDecodeResult{
			isError:  true,
			httpCode: http.StatusBadRequest,
			message:  "Request body must only contain a single JSON object",
		}
	}

	return jsonDecodeResult{
		isError: false,
	}
}

func responseAsJSON(w http.ResponseWriter, response interface{}, code int) {
	js, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_, err = w.Write(js)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}

}
