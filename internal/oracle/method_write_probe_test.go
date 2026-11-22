package oracle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Lower's probe for Compiler on #b5w3ycg: a write to a method's name through a cast castUnionUnwritten lets past.
func TestMethodWriteProbe(t *testing.T) {
	directory := "internal/oracle/testdata/method_write_probe"
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
		t.Logf("ROW %s: node %q exit %d | native exit %d %q %q | js exit %d %q", entry.Name(), strings.TrimSpace(string(truth.stdout)), truth.exitCode, sanitized.exitCode, strings.TrimSpace(string(sanitized.stdout)), probeFirstLine(string(sanitized.stderr)), js.exitCode, strings.TrimSpace(string(js.stdout)))
	}
}

func probeFirstLine(text string) string {
	for _, line := range strings.Split(text, "\n") {
		if strings.TrimSpace(line) != "" {
			return strings.TrimSpace(line)
		}
	}
	return ""
}
