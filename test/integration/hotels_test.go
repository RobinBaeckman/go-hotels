package integration

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
)

var (
	apiBaseURL string
	doc        *openapi3.T
)

func TestMain(m *testing.M) {
	apiBaseURL = getEnv("API_BASE_URL", "http://localhost:8080")
	waitForAPI(apiBaseURL + "/ready")

	// Load OpenAPI spec without needing *testing.T
	doc = mustLoadSpec()
	LoadAllRoutesFromSpec(doc)

	code := m.Run()
	PrintCoverageReport()

	// Check coverage threshold
	if err := AssertFullCoverage(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	os.Exit(code)
}

func TestCreateAndListHotels(t *testing.T) {
	// --- POST /hotels ---
	payload := `{
        "name": "Test Hotel",
        "city": "Göteborg",
        "stars": 4,
        "price_per_night": 123.45,
        "amenities": ["wifi", "breakfast"]
    }`

	req, _ := http.NewRequest(http.MethodPost, apiBaseURL+"/hotels", bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("POST request failed: %v", err)
	}
	defer func() {
		if err = resp.Body.Close(); err != nil {
			t.Fatalf("failed to close POST response body: %v", err)
		}
	}()
	validateResponse(t, doc, req, resp)

	// --- GET /hotels?city=Göteborg ---
	req, _ = http.NewRequest(http.MethodGet, apiBaseURL+"/hotels?city=Göteborg", nil)
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("GET request failed: %v", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			t.Fatalf("failed to close GET response body: %v", err)
		}
	}()
	validateResponse(t, doc, req, resp)
}

// Load OpenAPI spec without testing.T (for TestMain)
func mustLoadSpec() *openapi3.T {
	loader := openapi3.NewLoader()

	_, filename, _, _ := runtime.Caller(0)
	baseDir := filepath.Join(filepath.Dir(filename), "../../")
	specPath := filepath.Join(baseDir, "api/openapi.yaml")

	doc, err := loader.LoadFromFile(specPath)
	if err != nil {
		panic("failed to load OpenAPI spec: " + err.Error())
	}
	if err := doc.Validate(context.Background()); err != nil {
		panic("OpenAPI spec is invalid: " + err.Error())
	}
	return doc
}

func TestHealthAndReady(t *testing.T) {
	endpoints := []string{"/health", "/ready"}

	for _, path := range endpoints {
		req, _ := http.NewRequest(http.MethodGet, apiBaseURL+path, nil)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("%s request failed: %v", path, err)
		}
		defer func(resp *http.Response) {
			if err := resp.Body.Close(); err != nil {
				t.Fatalf("failed to close %s response body: %v", path, err)
			}
		}(resp)
		validateResponse(t, doc, req, resp)
	}
}
