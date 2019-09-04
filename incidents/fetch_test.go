package incidents

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetResponse(t *testing.T) {
	// failure case
	res, err := GetResponse("not_found")
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
	if res != nil {
		t.Errorf("Expected response, got nil")
	}

	// success case
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, World")
	}))
	defer ts.Close()
	res, err = GetResponse(ts.URL)
	if err != nil {
		t.Errorf("Expected nil, got %v\n", err)
	}
	if res == nil {
		t.Errorf("Expected response, got nil")
	}

}

func TestValidateResponse(t *testing.T) {
	// Create new http handler and serve a request using the handler
	handler := func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "<html><body>Hello World!</body></html>")
	}

	req := httptest.NewRequest("GET", "http://example.com/foo", nil)
	w := httptest.NewRecorder()
	handler(w, req)

	resp := w.Result()

	// statusCode non 200
	resp.StatusCode = 500
	err := ValidateResponse(resp)
	if err == nil {
		t.Errorf("Expected 500 error")
	}

	// content type
	resp.StatusCode = 200
	resp.Header["Content-Type"][0] = "text/html; charset=utf-8"
	err = ValidateResponse(resp)
	if err == nil {
		t.Errorf("Expected content-type mismatch error")
	}

	// content length
	resp.Header["Content-Type"][0] = "application/json"
	resp.Header["Content-Length"] = []string{"50"}
	err = ValidateResponse(resp)
	if err == nil {
		t.Errorf("Expected error, got no error")
	}
}
