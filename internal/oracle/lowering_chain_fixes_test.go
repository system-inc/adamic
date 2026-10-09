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

func TestLoweringChainStructuralError(t *testing.T) {
	t.Parallel()
	loweringChainAgreement(t, "internal/oracle/testdata/statements_small_stopped/structural_error.a")
}

func loweringChainAgreement(t *testing.T, fixture string) {
	t.Helper()
	path, err := filepath.Abs(filepath.Join(repository, fixture))
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
			t.Fatalf("%s: %s; Node %+v; backend %+v", mode, difference, truth, result)
		}
	}
	if report := leaksUncached(t, program, binary); report != "" {
		t.Fatal(report)
	}
}

func init() {
	for _, path := range []string{"internal/oracle/testdata/lowering_chain_namespace_catch.a", "internal/oracle/testdata/lowering_chain_tdz_catch.a", "internal/oracle/testdata/statements_small_stopped/structural_error.a"} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{path, true, false})
	}
}
func TestLoweringChainNamespaceCatch(t *testing.T) {
	t.Parallel()
	loweringChainAgreement(t, "internal/oracle/testdata/lowering_chain_namespace_catch.a")
}
func TestLoweringChainTDZCatch(t *testing.T) {
	t.Parallel()
	loweringChainAgreement(t, "internal/oracle/testdata/lowering_chain_tdz_catch.a")
}
