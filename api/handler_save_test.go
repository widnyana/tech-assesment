package api

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

var (
	newsJSON = `{"author":"bambang", "body": "pisang goreng"}`
)

func TestCreateNew(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/news", strings.NewReader(newsJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	createContent(w, req)

	resp := w.Result()
	body, _ := ioutil.ReadAll(resp.Body)

	fmt.Println(resp.StatusCode)
	fmt.Println(resp.Header.Get("Content-Type"))
	fmt.Println(string(body))
}
