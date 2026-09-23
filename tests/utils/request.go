package utils

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// BuildJSONRequest creates a new http.Request with the given body, url and verb.
func BuildJSONRequest(t *testing.T, body interface{}, url string, verb string) *http.Request {
	var (
		bodyBytes []byte
		err       error
	)

	// Check if body is of type []byte
	if b, ok := body.([]byte); ok {
		bodyBytes = b
	} else {
		bodyBytes, err = json.Marshal(&body)
		assert.NoError(t, err)
	}

	req := httptest.NewRequest(verb, url, bytes.NewBuffer(bodyBytes))

	return req
}

// StartTestServer creates a new http.Client, http.ServeMux and httptest.Server.
func StartTestServer() (*http.Client, *http.ServeMux, *httptest.Server) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	client := &http.Client{}

	return client, mux, server
}