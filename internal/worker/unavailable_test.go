package worker

import (
	"bytes"
	"regexp"
	"testing"
)

// A Node global Workers lacks is a use only where its spelling starts an identifier: a standard
// global that ends in one, like ArrayBuffer.isView, which every emitted module reads through
// adamicPlaceholderNarrow, is not a use. This is internal/javascript's nodeGlobalSpelling, the same
// rule for the files a Worker ships.
var nodeGlobalSpelling = regexp.MustCompile(`(^|[^A-Za-z0-9_$])(node:|process\.|Buffer\.)`)

// workersUnavailable is the first spelling in source that Workers can't run, or "".
func workersUnavailable(source []byte) string {
	if match := nodeGlobalSpelling.FindSubmatch(source); match != nil {
		return string(match[2])
	}
	for _, forbidden := range []string{"eval(", "new Function("} {
		if bytes.Contains(source, []byte(forbidden)) {
			return forbidden
		}
	}
	return ""
}

func TestWorkersUnavailableMatchesOnlyNodeGlobals(t *testing.T) {
	t.Parallel()
	for source, want := range map[string]string{
		"Buffer.from(text)":                       "Buffer.",
		"(process.env.HOME)":                      "process.",
		`import { readFileSync } from "node:fs";`: "node:",
		"return eval(text);":                      "eval(",
		"new Function('return 1')":                "new Function(",
		"!ArrayBuffer.isView(value)":              "",
		"adamicProcess.exit(0)":                   "",
	} {
		if got := workersUnavailable([]byte(source)); got != want {
			t.Errorf("workersUnavailable(%q) = %q, want %q", source, got, want)
		}
	}
}
