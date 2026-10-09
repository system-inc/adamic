package oracle

import (
	"bytes"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/native"
)

func init() {
	for _, name := range []string{"legal", "benchmark", "pathological"} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{"stage3/parser-next/stack/" + name + ".a", true, false})
	}
	// The stop depth varies with frame size and process stack layout.
	uncounted["stage3/parser-next/stack/pathological.a"] = true
}
func parserStackPath(t *testing.T, name string) string {
	t.Helper()
	p, err := filepath.Abs(filepath.Join(repository, "stage3/parser-next/stack", name+".a"))
	if err != nil {
		t.Fatal(err)
	}
	return p
}
func TestParserStackLegal(t *testing.T) {
	t.Parallel()
	path := parserStackPath(t, "legal")
	truth := onNode(t, path)
	if truth.exitCode != 0 || string(truth.stdout) != "859|1719\n" || len(truth.stderr) != 0 {
		t.Fatalf("Node %+v", truth)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	actual, binary := nativelyUncached(t, program)
	for backend, got := range map[string]run{"native": actual, "JavaScript": onJavaScriptBackend(t, program)} {
		if d := disagreement(truth, got); d != "" {
			t.Fatalf("%s %s", backend, d)
		}
	}
	if report := leaksUncached(t, program, binary); report != "" {
		t.Fatal(report)
	}
	t.Log("859 nested parentheses: source Node, native sanitizers/leaks and JavaScript print 859|1719")
}
func TestParserStackPathological(t *testing.T) {
	t.Parallel()
	path := parserStackPath(t, "pathological")
	truth := onNode(t, path)
	if truth.exitCode != 70 || !strings.Contains(string(truth.stderr), "RangeError: Maximum call stack size exceeded") {
		t.Fatalf("Node %+v", truth)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	actual, _ := nativelyUncached(t, program)
	want := "adamic: panic: RangeError: Maximum call stack size exceeded in parseNested\n"
	for backend, got := range map[string]run{"native": actual, "JavaScript": onJavaScriptBackend(t, program, path), "release": released(t, program)} {
		if got.exitCode != 70 || len(got.stdout) != 0 || string(got.stderr) != want {
			t.Errorf("%s %+v, want %q", backend, got, want)
		}
	}
	t.Logf("source Node outcome: exit %d, stderr %q; native named guard pinned separately", truth.exitCode, truth.stderr)
}

// Native-only: Node itself cannot start safely under these stack limits.
func TestParserStackTinyLimit(t *testing.T) {
	t.Parallel()
	program, err := lowered(t, parserStackPath(t, "pathological"))
	if err != nil {
		t.Fatal(err)
	}
	for _, sanitized := range []bool{false, true} {
		binary := filepath.Join(t.TempDir(), "native")
		if err := native.Build(native.C(program), binary, native.Options{Sanitize: sanitized}); err != nil {
			t.Fatal(err)
		}
		for _, limit := range []string{"64", "128"} {
			got := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=0"}, "/bin/sh", "-c", `ulimit -s "$1" && exec "$0"`, binary, limit)
			want := "adamic: panic: RangeError: Maximum call stack size exceeded in parseNested\n"
			if got.exitCode != 70 || string(got.stderr) != want || len(got.stdout) != 0 {
				t.Errorf("%s KiB sanitized=%t: exit %d stderr %.500q", limit, sanitized, got.exitCode, got.stderr)
			}
		}
	}
}

func parserStackWithoutGuard(t *testing.T, source string) string {
	t.Helper()
	pattern := regexp.MustCompile(`(?m)^\s*ADAMIC_CHECK_STACK_NAMED\([^\n]*\);\n`)
	changed := pattern.ReplaceAllString(source, "\n")
	if changed == source {
		t.Fatal("removed no guards")
	}
	return changed
}
func TestParserStackGuardMutant(t *testing.T) {
	t.Parallel()
	program, err := lowered(t, parserStackPath(t, "pathological"))
	if err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(t.TempDir(), "mutant")
	if err := native.Build(parserStackWithoutGuard(t, native.C(program)), binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	actual := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=0"}, "/bin/sh", "-c", `ulimit -s 1024 && exec "$0"`, binary)
	if actual.exitCode == 0 || actual.exitCode == 70 || !bytes.Contains(actual.stderr, []byte("AddressSanitizer: stack-overflow")) {
		t.Fatalf("guard removal not caught by stack sanitizer: exit %d stderr %.500q", actual.exitCode, actual.stderr)
	}
	t.Logf("removed guard compiles and executes: exit %d, AddressSanitizer: stack-overflow", actual.exitCode)
}

// Not parallel: process timing requires alternating quiet release runs.
func TestParserStackGuardCost(t *testing.T) {
	if os.Getenv("ADAMIC_MEASURE_STACK_GUARD") == "" {
		t.Skip("set ADAMIC_MEASURE_STACK_GUARD=1 for the local cost record")
	}
	for _, name := range []string{"stack/benchmark", "speculation/named-lookahead", "speculation/named-tryparse", "speculation/arrow-tryparse", "speculation/truthiness-rewind"} {
		path, err := filepath.Abs(filepath.Join(repository, "stage3/parser-next", name+".a"))
		if err != nil {
			t.Fatal(err)
		}
		program, err := lowered(t, path)
		if err != nil {
			t.Fatal(err)
		}
		truth := onNode(t, path)
		if truth.exitCode != 0 {
			t.Fatal(truth)
		}
		source := native.C(program)
		guarded := filepath.Join(t.TempDir(), "guarded")
		unguarded := filepath.Join(t.TempDir(), "unguarded")
		for binary, code := range map[string]string{guarded: source, unguarded: parserStackWithoutGuard(t, source)} {
			if err := native.Build(code, binary, native.Options{}); err != nil {
				t.Fatal(err)
			}
		}
		var samples [2][]time.Duration
		for round := 0; round < 40; round++ {
			for offset := 0; offset < 2; offset++ {
				which := (round + offset) % 2
				binary := []string{guarded, unguarded}[which]
				start := time.Now()
				got := execute(t, binary)
				elapsed := time.Since(start)
				if d := disagreement(truth, got); d != "" {
					t.Fatal(d)
				}
				if round >= 4 {
					samples[which] = append(samples[which], elapsed)
				}
			}
		}
		var medians [2]time.Duration
		for which := range samples {
			sort.Slice(samples[which], func(i, j int) bool { return samples[which][i] < samples[which][j] })
			medians[which] = (samples[which][17] + samples[which][18]) / 2
		}
		t.Logf("%s: 36 measured paired release process runs after 4 warmups; guarded median %s, removed median %s, change %.2f%%; includes process launch", name, medians[0], medians[1], 100*(float64(medians[0])/float64(medians[1])-1))
	}
}
