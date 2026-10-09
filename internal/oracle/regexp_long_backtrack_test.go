package oracle

import (
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/childguard"
	"github.com/system-inc/adamic/internal/native"
)

// The same fixture registry shape, counted by the ordinary counts gate, but run
// below with a three-minute CPU budget and a separate generous hang cap.
var slowRegExpFixtures = []struct {
	path            string
	lowers, checked bool
}{{"internal/oracle/testdata/regexp_native_long_backtrack.a", true, false}}

// Not parallel: the long failing search deliberately exercises backtracking;
// avoid multiplying the expensive backtracking and sanitizer work.
func TestRegExpLongBacktrackNode(t *testing.T) {
	path, err := filepath.Abs(filepath.Join(repository, slowRegExpFixtures[0].path))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	want := onNode(t, path)
	if want.exitCode != 0 {
		t.Fatalf("Node did not finish: %q", want.stderr)
	}
	if got := onJavaScriptBackend(t, program); disagreement(want, got) != "" {
		t.Fatalf("JavaScript backend: %s", disagreement(want, got))
	}
	sanitized := filepath.Join(t.TempDir(), "sanitized")
	if err := native.Build(native.C(program), sanitized, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	got := longRegExpRun(t, []string{"ASAN_OPTIONS=detect_leaks=0"}, sanitized)
	if difference := disagreement(want, got); difference != "" {
		t.Fatalf("native: %s Node=%q native=%q stderr=%q", difference, want.stdout, got.stdout, got.stderr)
	}
	release := filepath.Join(t.TempDir(), "release")
	if err := native.Build(native.C(program), release, native.Options{}); err != nil {
		t.Fatal(err)
	}
	if got := longRegExpRun(t, nil, release); disagreement(want, got) != "" {
		t.Fatalf("release build: %s", disagreement(want, got))
	}
	var leaked run
	switch runtime.GOOS {
	case "linux":
		leaked = longRegExpRun(t, []string{"ASAN_OPTIONS=detect_leaks=1"}, sanitized)
	case "darwin":
		leaked = longRegExpRun(t, nil, "leaks", "--atExit", "--", release)
	default:
		t.Fatalf("no leak oracle for %s", runtime.GOOS)
	}
	if leaked.exitCode != 0 {
		t.Fatalf("leaks: exit %d\n%s\n%s", leaked.exitCode, leaked.stderr, leaked.stdout)
	}
}

func longRegExpRun(t *testing.T, environment []string, name string, arguments ...string) run {
	t.Helper()
	budget := 3 * time.Minute
	// leaks is an external tool; the release child was CPU-checked above.
	if name == "leaks" {
		budget = 0
	}
	result, cpu, err := regExpCPUCommand(environment, name, arguments, childguard.Options{}, budget)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%s %v: CPU %s, exit %d", filepath.Base(name), environment, cpu, result.exitCode)
	rememberRun(t, result)
	return result
}
