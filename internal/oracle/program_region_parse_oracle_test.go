package oracle

import (
	"path/filepath"
	"runtime"
	"testing"
)

func programParseOracle(t *testing.T, enabled bool) {
	t.Helper()
	if runtime.GOOS != "linux" {
		t.Skip("Linux instrumented parse witness")
	}
	args := programParseArguments(t)
	absolute, err := filepath.Abs("testdata/program_region/parse.a")
	if err != nil {
		t.Fatal(err)
	}
	truth := execute(t, "node", append([]string{"--disable-warning=ExperimentalWarning", "../../oracle/node.mjs", absolute}, args...)...)
	if truth.exitCode != 0 {
		t.Fatalf("source Node: %d %s", truth.exitCode, truth.stderr)
	}
	js := filepath.Join(programParseLower(t, enabled), "program.mjs")
	actual := execute(t, "node", append([]string{"--disable-warning=ExperimentalWarning", "../../oracle/node.mjs", js}, args...)...)
	if difference := disagreement(truth, actual); difference != "" {
		t.Fatalf("JavaScript: %s", difference)
	}
	binary := filepath.Join(programParseBuildKind(t, enabled, true), "parse")
	actual = executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=1", "UBSAN_OPTIONS=halt_on_error=1"}, binary, args...)
	actual.stdout = graphRSS.ReplaceAll(actual.stdout, nil)
	if difference := disagreement(truth, actual); difference != "" {
		t.Fatalf("sanitized parse: %s", difference)
	}
}
func TestProgramRegionParseOnAgreesWithNode(t *testing.T) { t.Parallel(); programParseOracle(t, true) }
func TestProgramRegionParseOffAgreesWithNode(t *testing.T) {
	t.Parallel()
	programParseOracle(t, false)
}
