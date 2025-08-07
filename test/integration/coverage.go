package integration

import (
	"fmt"
	"os"
	"path/filepath"
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
	var sb strings.Builder
	total := 0
	covered := 0

	sb.WriteString("### 📊 Endpoint Coverage Report\n\n")

	for path, methods := range allRoutes {
		for method := range methods {
			total++
			if coveredRoutes[path][method] {
				sb.WriteString(fmt.Sprintf("✅ `%s %s`\n", method, path))
				covered++
			} else {
				sb.WriteString(fmt.Sprintf("❌ `%s %s`\n", method, path))
			}
		}
	}

	percentage := float64(covered) / float64(total) * 100
	sb.WriteString(fmt.Sprintf("\n**Covered %d/%d endpoints (%.1f%%)**\n", covered, total, percentage))

	fmt.Println(sb.String())

	if summary := os.Getenv("GITHUB_STEP_SUMMARY"); summary != "" {
		if !filepath.IsAbs(summary) {
			fmt.Fprintf(os.Stderr, "⚠️ Refusing to write summary to non-absolute path: %s\n", summary)
			return
		}
		f, err := os.OpenFile(summary, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600) // #nosec G304
		if err == nil {
			_, _ = f.WriteString(sb.String())
			if cerr := f.Close(); cerr != nil {
				fmt.Fprintf(os.Stderr, "⚠️ Failed to close summary file: %v\n", cerr)
			}
		} else {
			fmt.Fprintf(os.Stderr, "⚠️ Failed to open summary file: %v\n", err)
		}
	}
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
