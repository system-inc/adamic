package prettier

import (
	"os"
	"strings"
	"testing"
	"time"
)

// TestCost measures formatting throughput per byte rather than per file.
//
// Per-file cost is not a property of the formatter, it is a property of the sample: a corpus of
// 9.5 KB files will report roughly twice the per-file cost of a corpus of 4.3 KB files while the
// engine does exactly the same work per byte. Two honest measurements of the same engine can
// therefore disagree by 2x with neither being wrong, which is what happened here.
//
//	COHERE_COST_LIST=/path/to/list go test ./internal/prettier/ -run TestCost -v
func TestCost(t *testing.T) {
	t.Parallel()

	listPath := os.Getenv("COHERE_COST_LIST")
	if listPath == "" {
		t.Skip("set COHERE_COST_LIST")
	}
	listBytes, err := os.ReadFile(listPath)
	if err != nil {
		t.Fatalf("reading the list: %v", err)
	}

	constructionStart := time.Now()
	engine := newTestEngine(t)
	construction := time.Since(constructionStart)

	var totalBytes int
	var formatted int
	var formatting time.Duration
	for _, path := range strings.Split(strings.TrimSpace(string(listBytes)), "\n") {
		path = strings.TrimSpace(path)
		if path == "" || !engine.Handles(path) {
			continue
		}
		source, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		start := time.Now()
		if _, err := engine.Format(path, string(source)); err != nil {
			t.Errorf("%s: %v", path, err)
			continue
		}
		formatting += time.Since(start)
		totalBytes += len(source)
		formatted++
	}

	if formatted == 0 {
		t.Fatal("measured no files, so this run says nothing")
	}
	t.Logf("construction %v (once)", construction)
	t.Logf("files %d, bytes %d, mean file %d bytes", formatted, totalBytes, totalBytes/formatted)
	t.Logf("formatting %v total, %v per file, %.1f KB/s",
		formatting, formatting/time.Duration(formatted),
		float64(totalBytes)/1024/formatting.Seconds())
}
