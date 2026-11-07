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
func parserStackPath(test *testing.T, name string) string {
	test.Helper()
	p, absolutePathError := filepath.Abs(filepath.Join(repository, "stage3/parser-next/stack", name+".a"))
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	return p
}
func TestParserStackLegal(test *testing.T) {
	test.Parallel()
	path := parserStackPath(test, "legal")
	truth := onSliceNode(test, path)
	if truth.exitCode != 0 || string(truth.stdout) != "859|1719\n" || len(truth.stderr) != 0 {
		test.Fatalf("Node %+v", truth)
	}
	program, loweredParserError := loweredParser(test, path)
	if loweredParserError != nil {
		test.Fatal(loweredParserError)
	}
	actual, binary := nativelyUncached(test, program)
	for backend, got := range map[string]run{"native": actual, "JavaScript": onSliceJavaScript(test, program)} {
		if d := disagreement(truth, got); d != "" {
			test.Fatalf("%s %s", backend, d)
		}
	}
	if report := leaksUncached(test, program, binary); report != "" {
		test.Fatal(report)
	}
	test.Log("859 nested parentheses: source Node, native sanitizers/leaks and JavaScript print 859|1719")
}
func TestParserStackPathological(test *testing.T) {
	test.Parallel()
	path := parserStackPath(test, "pathological")
	truth := onSliceNode(test, path)
	if truth.exitCode != 70 || !strings.Contains(string(truth.stderr), "RangeError: Maximum call stack size exceeded") {
		test.Fatalf("Node %+v", truth)
	}
	program, loweredParserError := loweredParser(test, path)
	if loweredParserError != nil {
		test.Fatal(loweredParserError)
	}
	actual, _ := nativelyUncached(test, program)
	want := "adamic: panic: RangeError: Maximum call stack size exceeded in parseNested\n"
	for backend, got := range map[string]run{
		"native":     actual,
		"JavaScript": onSliceJavaScript(test, program),
		"release":    releasedUncached(test, program),
	} {
		if got.exitCode != 70 || len(got.stdout) != 0 || string(got.stderr) != want {
			test.Errorf("%s %+v, want %q", backend, got, want)
		}
	}
	test.Logf(
		"source Node outcome: exit %d, stderr %q; native named guard pinned separately",
		truth.exitCode,
		truth.stderr,
	)
}

// Native-only: Node itself cannot start safely under these stack limits.
func TestParserStackTinyLimit(test *testing.T) {
	test.Parallel()
	program, loweredParserError := loweredParser(test, parserStackPath(test, "pathological"))
	if loweredParserError != nil {
		test.Fatal(loweredParserError)
	}
	for _, sanitized := range []bool{false, true} {
		binary := filepath.Join(test.TempDir(), "native")
		if buildError := native.Build(
			native.C(program),
			binary,
			native.Options{Sanitize: sanitized},
		); buildError != nil {
			test.Fatal(buildError)
		}
		for _, limit := range []string{"64", "128"} {
			got := executeWith(
				test,
				[]string{"ASAN_OPTIONS=detect_leaks=0"},
				"/bin/sh",
				"-c",
				`ulimit -s "$1" && exec "$0"`,
				binary,
				limit,
			)
			want := "adamic: panic: RangeError: Maximum call stack size exceeded in parseNested\n"
			if got.exitCode != 70 || string(got.stderr) != want || len(got.stdout) != 0 {
				test.Errorf("%s KiB sanitized=%t: exit %d stderr %.500q", limit, sanitized, got.exitCode, got.stderr)
			}
		}
	}
}

func parserStackWithoutGuard(test *testing.T, source string) string {
	test.Helper()
	pattern := regexp.MustCompile(`(?m)^\s*ADAMIC_CHECK_STACK_NAMED\([^\n]*\);\n`)
	changed := pattern.ReplaceAllString(source, "\n")
	if changed == source {
		test.Fatal("removed no guards")
	}
	return changed
}
func TestParserStackGuardMutant(test *testing.T) {
	test.Parallel()
	program, loweredParserError := loweredParser(test, parserStackPath(test, "pathological"))
	if loweredParserError != nil {
		test.Fatal(loweredParserError)
	}
	binary := filepath.Join(test.TempDir(), "mutant")
	if buildError := native.Build(
		parserStackWithoutGuard(test, native.C(program)),
		binary,
		native.Options{Sanitize: true},
	); buildError != nil {
		test.Fatal(buildError)
	}
	actual := executeWith(
		test,
		[]string{"ASAN_OPTIONS=detect_leaks=0"},
		"/bin/sh",
		"-c",
		`ulimit -s 1024 && exec "$0"`,
		binary,
	)
	if actual.exitCode == 0 ||
		actual.exitCode == 70 ||
		!bytes.Contains(actual.stderr, []byte("AddressSanitizer: stack-overflow")) {
		test.Fatalf(
			"guard removal not caught by stack sanitizer: exit %d stderr %.500q",
			actual.exitCode,
			actual.stderr,
		)
	}
	test.Logf("removed guard compiles and executes: exit %d, AddressSanitizer: stack-overflow", actual.exitCode)
}

// Not parallel: process timing requires alternating quiet release runs.
func TestParserStackGuardCost(test *testing.T) {
	if os.Getenv("ADAMIC_MEASURE_STACK_GUARD") == "" {
		test.Skip("set ADAMIC_MEASURE_STACK_GUARD=1 for the local cost record")
	}
	for _, name := range []string{
		"stack/benchmark",
		"speculation/named-lookahead",
		"speculation/named-tryparse",
		"speculation/arrow-tryparse",
		"speculation/truthiness-rewind",
	} {
		path, absolutePathError := filepath.Abs(filepath.Join(repository, "stage3/parser-next", name+".a"))
		if absolutePathError != nil {
			test.Fatal(absolutePathError)
		}
		program, loweredParserError := loweredParser(test, path)
		if loweredParserError != nil {
			test.Fatal(loweredParserError)
		}
		truth := onSliceNode(test, path)
		if truth.exitCode != 0 {
			test.Fatal(truth)
		}
		source := native.C(program)
		guarded := filepath.Join(test.TempDir(), "guarded")
		unguarded := filepath.Join(test.TempDir(), "unguarded")
		for binary, code := range map[string]string{guarded: source, unguarded: parserStackWithoutGuard(test, source)} {
			if buildError := native.Build(code, binary, native.Options{}); buildError != nil {
				test.Fatal(buildError)
			}
		}
		var samples [2][]time.Duration
		for round := 0; round < 40; round++ {
			for offset := 0; offset < 2; offset++ {
				which := (round + offset) % 2
				binary := []string{guarded, unguarded}[which]
				start := time.Now()
				got := execute(test, binary)
				elapsed := time.Since(start)
				if d := disagreement(truth, got); d != "" {
					test.Fatal(d)
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
		test.Logf(
			"%s: 36 measured paired release process runs after 4 warmups; guarded median %s, removed median %s, change %.2f%%; includes process launch",
			name,
			medians[0],
			medians[1],
			100*(float64(medians[0])/float64(medians[1])-1),
		)
	}
}
