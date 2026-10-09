package oracle

import (
	"path/filepath"
	"testing"
)

func init() {
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/lowering_chain_regex_fields.a", true, false})
}

func TestLoweringChainRegexFields(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/lowering_chain_regex_fields.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	truth := onNode(t, path)
	sanitized, binary := nativelyUncached(t, program)
	for mode, result := range map[string]run{"JavaScript": onJavaScriptBackend(t, program), "sanitized": sanitized, "release": releasedUncached(t, program)} {
		if difference := disagreement(truth, result); difference != "" {
			t.Fatalf("%s: %s", mode, difference)
		}
	}
	if report := leaksUncached(t, program, binary); report != "" {
		t.Fatal(report)
	}
}
