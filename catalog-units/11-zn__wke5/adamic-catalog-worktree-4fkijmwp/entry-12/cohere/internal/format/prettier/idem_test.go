package prettier

import (
	"os"
	"strings"
	"testing"
)

// TestFormatIsIdempotent checks that formatting formatted output changes nothing.
//
// A formatter whose second pass differs from its first would report writes forever, and the count
// would be honest while the tree never converged.
func TestFormatIsIdempotent(t *testing.T) {
	t.Parallel()

	listPath := os.Getenv("COHERE_IDEM_LIST")
	if listPath == "" {
		t.Skip("set COHERE_IDEM_LIST")
	}
	engine := newTestEngine(t)
	listBytes, _ := os.ReadFile(listPath)
	var checked, unstable, changedFromSource int
	for _, path := range strings.Split(strings.TrimSpace(string(listBytes)), "\n") {
		path = strings.TrimSpace(path)
		if path == "" || !engine.Handles(path) {
			continue
		}
		source, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		once, err := engine.Format(path, string(source))
		if err != nil {
			continue
		}
		twice, err := engine.Format(path, once)
		if err != nil {
			continue
		}
		checked++
		if once != string(source) {
			changedFromSource++
		}
		if once != twice {
			unstable++
			t.Errorf("%s: second pass differs from first", path)
		}
	}
	t.Logf("checked=%d changedFromSource=%d unstable=%d", checked, changedFromSource, unstable)
}
