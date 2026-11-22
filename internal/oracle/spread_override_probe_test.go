package oracle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Lower's probe of #sx8pbw1 (B): each program's lowering stop, or Node beside every backend.
func TestSpreadOverrideProbe(t *testing.T) {
	directory := "internal/oracle/testdata/spread_override_probe"
	entries, err := os.ReadDir(filepath.Join(repository, directory))
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		absolute, err := filepath.Abs(filepath.Join(repository, directory, entry.Name()))
		if err != nil {
			t.Fatal(err)
		}
		truth := onNode(t, absolute)
		program, err := lowered(t, absolute)
		if err != nil {
			t.Logf("ROW %s: node %q exit %d | STOP %v", entry.Name(), strings.TrimSpace(string(truth.stdout)), truth.exitCode, err)
			continue
		}
		sanitized, _ := natively(t, program)
		js := onJavaScriptBackend(t, program)
		t.Logf("ROW %s: node %q exit %d | native exit %d %q %q | js exit %d %q", entry.Name(), strings.TrimSpace(string(truth.stdout)), truth.exitCode, sanitized.exitCode, strings.TrimSpace(string(sanitized.stdout)), firstLine(string(sanitized.stderr)), js.exitCode, strings.TrimSpace(string(js.stdout)))
	}
}

func firstLine(text string) string {
	for _, line := range strings.Split(text, "\n") {
		if strings.TrimSpace(line) != "" {
			return strings.TrimSpace(line)
		}
	}
	return ""
}
