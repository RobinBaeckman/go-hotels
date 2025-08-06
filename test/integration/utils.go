package integration

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers/legacy"
)

// Get environment variable with fallback
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// Wait until the API is ready
func waitForAPI(url string) {
	client := &http.Client{Timeout: 2 * time.Second}
	for i := 0; i < 30; i++ {
		resp, err := client.Get(url)
		if err == nil && resp.StatusCode == http.StatusOK {
			_ = resp.Body.Close()
			return
		}
		time.Sleep(1 * time.Second)
	}
	panic(fmt.Sprintf("API at %s not ready after 30s", url))
}

// Validate a real HTTP response against the OpenAPI spec
func validateResponse(t *testing.T, doc *openapi3.T, req *http.Request, resp *http.Response) {
	t.Helper()

	router, err := legacy.NewRouter(doc)
	if err != nil {
		t.Fatalf("failed to create router: %v", err)
	}

	// Match route using full request
	route, pathParams, err := router.FindRoute(req)
	if err != nil {
		t.Fatalf("❌ Failed to find route %s %s: %v", req.Method, req.URL.Path, err)
	}

	RecordCoverage(req.Method, req.URL.Path)

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}
	if err := resp.Body.Close(); err != nil {
		t.Fatalf("failed to close response body: %v", err)
	}
	resp.Body = io.NopCloser(bytes.NewReader(bodyBytes))

	input := &openapi3filter.ResponseValidationInput{
		RequestValidationInput: &openapi3filter.RequestValidationInput{
			Request:    req,
			PathParams: pathParams,
			Route:      route,
		},
		Status:  resp.StatusCode,
		Header:  resp.Header,
		Body:    io.NopCloser(bytes.NewReader(bodyBytes)),
		Options: &openapi3filter.Options{},
	}

	if err := openapi3filter.ValidateResponse(context.Background(), input); err != nil {
		t.Fatalf("response validation failed: %v", err)
	}
}
