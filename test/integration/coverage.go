package integration

import (
	"fmt"
	"strings"
	"sync"

	"github.com/getkin/kin-openapi/openapi3"
)

var (
	mu            sync.Mutex
	coveredRoutes = make(map[string]map[string]bool) // path -> method -> true
	allRoutes     = make(map[string]map[string]bool) // loaded from OpenAPI
)

func RecordCoverage(method, path string) {
	mu.Lock()
	defer mu.Unlock()

	method = normalize(method)
	if _, ok := coveredRoutes[path]; !ok {
		coveredRoutes[path] = make(map[string]bool)
	}
	coveredRoutes[path][method] = true
}

func LoadAllRoutesFromSpec(doc *openapi3.T) {
	for path, pathItem := range doc.Paths.Map() {
		allRoutes[path] = make(map[string]bool)
		for method := range pathItem.Operations() {
			allRoutes[path][normalize(method)] = false
		}
	}
}

func PrintCoverageReport() {
	fmt.Println("=== Endpoint Coverage Report ===")
	total := 0
	covered := 0

	for path, methods := range allRoutes {
		for method := range methods {
			total++
			if coveredRoutes[path][method] {
				fmt.Printf("✅ %s %s\n", method, path)
				covered++
			} else {
				fmt.Printf("❌ %s %s\n", method, path)
			}
		}
	}
	fmt.Printf("\nCovered %d/%d endpoints (%.1f%%)\n", covered, total, float64(covered)/float64(total)*100)
}

func normalize(m string) string {
	// Normalize to uppercase (just in case)
	return strings.ToUpper(m)
}

func GetUncoveredRoutes() []string {
	var uncovered []string
	for path, methods := range allRoutes {
		for method := range methods {
			if !coveredRoutes[path][method] {
				uncovered = append(uncovered, fmt.Sprintf("%s %s", method, path))
			}
		}
	}
	return uncovered
}

func AssertFullCoverage() error {
	uncovered := GetUncoveredRoutes()
	if len(uncovered) == 0 {
		return nil
	}

	var b strings.Builder
	b.WriteString("❌ Integration test coverage incomplete. Uncovered endpoints:\n")
	for _, route := range uncovered {
		b.WriteString("  - " + route + "\n")
	}
	return fmt.Errorf("%s", b.String())
}
