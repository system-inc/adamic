package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/toolchain"
)

// TestClassifyCorpus locks the skip rules, including negative parse and early error. A runner that
// ignores a negative expectation classifies those fixtures as attempted, and this fails.
func TestClassifyCorpus(test *testing.T) {
	test.Parallel()
	matches, globError := filepath.Glob("testdata/corpus/*.js")
	if globError != nil {
		test.Fatal(globError)
	}
	if len(matches) == 0 {
		test.Fatal("no corpus fixtures")
	}
	for _, path := range matches {
		test.Run(filepath.Base(path), func(test *testing.T) {
			test.Parallel()
			source, readError := os.ReadFile(path)
			if readError != nil {
				test.Fatal(readError)
			}
			want, readError := os.ReadFile(strings.TrimSuffix(path, ".js") + ".want")
			if readError != nil {
				test.Fatal(readError)
			}
			lines := strings.Split(strings.TrimSpace(string(want)), "\n")
			got := classify(filepath.Base(path), string(source), false)
			kind := "attempted"
			if got.Skip != "" {
				kind = "skipped"
			}
			if kind != lines[0] {
				test.Errorf("%s: kind %s, want %s (skip %q)", path, kind, lines[0], got.Skip)
			}
			if kind == "skipped" && !strings.Contains(got.Skip, lines[1]) {
				test.Errorf("%s: skip %q, want it to contain %q", path, got.Skip, lines[1])
			}
			if kind == "attempted" && len(lines) > 1 {
				gotPhase := got.NegativePhase + " " + got.NegativeType
				if strings.TrimSpace(gotPhase) != lines[1] {
					test.Errorf("%s: negative %q, want %q", path, gotPhase, lines[1])
				}
			}
			if kind == "attempted" && !strings.Contains(got.Program, "function assertSameValue") {
				test.Errorf("%s: attempted program has no prelude", path)
			}
		})
	}
}

// TestVerdictCorpus locks the comparison. A runner that counts a native failure as a pass fails
// native_fails.json. A runner that ignores a negative expectation fails negative_observed.json,
// because both sides exited non-zero and an ordinary comparison only passes when both exit 0.
func TestVerdictCorpus(test *testing.T) {
	test.Parallel()
	matches, globError := filepath.Glob("testdata/verdicts/*.json")
	if globError != nil {
		test.Fatal(globError)
	}
	if len(matches) == 0 {
		test.Fatal("no verdict fixtures")
	}
	for _, path := range matches {
		test.Run(filepath.Base(path), func(test *testing.T) {
			test.Parallel()
			data, readError := os.ReadFile(path)
			if readError != nil {
				test.Fatal(readError)
			}
			var fixture struct {
				verdictInput
				Name string `json:"name"`
				Want string `json:"want"`
			}
			if decodeError := json.Unmarshal(data, &fixture); decodeError != nil {
				test.Fatal(decodeError)
			}
			got := decide(fixture.verdictInput)
			if string(got.Kind) != fixture.Want {
				test.Errorf("%s (%s): got %s (%s), want %s", path, fixture.Name, got.Kind, got.Reason, fixture.Want)
			}
		})
	}
}

func TestNormalizeReason(test *testing.T) {
	test.Parallel()
	cases := []struct {
		stderr string
		want   string
	}{
		{"adamic: prog.ts:1:14: stage 0 can't lower a rest parameter yet\n", "not yet: a rest parameter"},
		{"adamic: /tmp/work/program.ts:1:1: Adamic 0.1 refuses var; use const or let\n", "refuses var"},
		{
			"program.ts:1:12: error TS7006: Parameter 'x' implicitly has an 'any' type.\n",
			"error TS7006: Parameter '…' implicitly has an '…' type.",
		},
	}
	for _, testCase := range cases {
		test.Run(testCase.stderr, func(test *testing.T) {
			test.Parallel()
			kind, reason := compileClass(testCase.stderr, 1, false)
			if kind != "refused" {
				test.Errorf("%q: kind %s, want refused", testCase.stderr, kind)
			}
			if reason != testCase.want {
				test.Errorf("normalize %q:\n got %q\nwant %q", testCase.stderr, reason, testCase.want)
			}
		})
	}
	kind, _ := compileClass("panic: something\ngoroutine 1 [running]:\n", 2, false)
	if kind != "crashed" {
		test.Errorf("compiler panic: kind %s, want crashed", kind)
	}
}

func TestRewriteHarnessCalls(test *testing.T) {
	test.Parallel()
	source := "const text = \"assert.sameValue(1, 2)\";\nassert.sameValue(1, 1);\nassert.throws(TypeError, function () {});\nassert.notSameValue(0, -0);\n"
	got := rewriteHarnessCalls(source)
	if strings.Contains(got, "\"assertSameValue") {
		test.Fatalf("rewrote inside a string:\n%s", got)
	}
	if !strings.Contains(got, "assertSameValue(1, 1);") {
		test.Fatalf("missed sameValue:\n%s", got)
	}
	if !strings.Contains(got, "assertThrows(\"TypeError\", function () {});") {
		test.Fatalf("missed throws:\n%s", got)
	}
	if !strings.Contains(got, "assertNotSameValue(0, -0);") {
		test.Fatalf("missed notSameValue:\n%s", got)
	}
	template := "assert.sameValue(`${'assert.sameValue'}`, 'x');\n"
	rewritten := rewriteHarnessCalls(template)
	if rewritten != "assertSameValue(`${'assert.sameValue'}`, 'x');\n" {
		test.Fatalf("template: %s", rewritten)
	}
	nested := "assert.sameValue(`${assert.sameValue(1, 1)}`, 'x');\n"
	rewritten = rewriteHarnessCalls(nested)
	if !strings.Contains(rewritten, "assertSameValue(`${assertSameValue(1, 1)}`, 'x');") {
		test.Fatalf("nested call: %s", rewritten)
	}
}

func TestFrontmatterShapes(test *testing.T) {
	test.Parallel()
	source := `/*---
description: >
    several
    lines
features:
  - String.prototype.padStart
  - arrow-function
flags: [onlyStrict, generated]
includes: [compareArray.js]
negative:
  phase: runtime
  type: TypeError
---*/
assert(true);
`
	got := parseFrontmatter(source)
	if !got.Present {
		test.Fatal("frontmatter not found")
	}
	if strings.Join(got.Features, ",") != "String.prototype.padStart,arrow-function" {
		test.Fatalf("features %v", got.Features)
	}
	if strings.Join(got.Flags, ",") != "onlyStrict,generated" {
		test.Fatalf("flags %v", got.Flags)
	}
	if strings.Join(got.Includes, ",") != "compareArray.js" {
		test.Fatalf("includes %v", got.Includes)
	}
	if got.NegativePhase != "runtime" || got.NegativeType != "TypeError" {
		test.Fatalf("negative %s %s", got.NegativePhase, got.NegativeType)
	}
	classified := classify("built-ins/String/x.js", source, false)
	if classified.Skip != "" {
		test.Fatalf("a padStart test was skipped: %s", classified.Skip)
	}
}

// TestMiniRunner runs the real command path over a three-file checkout: one test stage 0 and Node
// both pass, one negative parse test that must be skipped, and one `var` test the compiler refuses.
// Counting that refusal as a pass fails this.
func TestMiniRunner(test *testing.T) {
	test.Parallel()
	// A sandbox declares its tools, so one missing there is a failure, never a skip that passes having run nothing.
	for _, tool := range []string{"clang", "node"} {
		if _, pathError := toolchain.Path(tool); pathError != nil {
			if toolchain.Sandboxed() {
				test.Fatal(pathError)
			}
			test.Skip(tool, "not found:", pathError)
		}
	}
	work := test.TempDir()
	engine, prepareError := prepare(test, "testdata/mini", work)
	if prepareError != nil {
		test.Fatal(prepareError)
	}
	passed, runFilterError := engine.runFilter("pass", 0, false)
	if runFilterError != nil {
		test.Fatal(runFilterError)
	}
	if passed.Pass != 1 || passed.Fail != 0 || passed.Refused != 0 || passed.Crashed != 0 {
		test.Fatalf(
			"pass fixture: %+v reasons fail=%v refused=%v crash=%v",
			passed.Pass,
			passed.FailReasons,
			passed.RefusalReasons,
			passed.CrashReasons,
		)
	}
	skipped, runFilterError := engine.runFilter("skip", 0, false)
	if runFilterError != nil {
		test.Fatal(runFilterError)
	}
	if skipped.Skipped != 1 || skipped.Pass != 0 {
		test.Fatalf(
			"negative fixture: skipped %d pass %d reasons %v",
			skipped.Skipped,
			skipped.Pass,
			skipped.SkipReasons,
		)
	}
	refused, runFilterError := engine.runFilter("refuse", 0, false)
	if runFilterError != nil {
		test.Fatal(runFilterError)
	}
	if refused.Refused != 1 || refused.Pass != 0 {
		test.Fatalf("var fixture: refused %d pass %d reasons %v", refused.Refused, refused.Pass, refused.RefusalReasons)
	}
	if len(refused.RefusalReasons) == 0 || !strings.Contains(refused.RefusalReasons[0].Reason, "refuses var") {
		test.Fatalf("var fixture reason: %+v", refused.RefusalReasons)
	}
}

// A large sort previously reached clang with its C cut at 256 KiB, falsely blaming the emitter.
func TestLargeCompilerOutputIsComplete(test *testing.T) {
	test.Parallel()
	root := test.TempDir()
	if createDirectoryError := os.MkdirAll(filepath.Join(root, "test"), 0755); createDirectoryError != nil {
		test.Fatal(createDirectoryError)
	}
	// Each line is about 442 bytes of C, so 1000 lines are about 455 KB, 1.7 times the old limit: room for a
	// leaner emitter, without the 800 KB main of 1800 lines, which made this test 23 s alone on one core.
	source := strings.Repeat("assert.sameValue(1, 1);\n", 1000)
	if writeError := os.WriteFile(filepath.Join(root, "test", "large.js"), []byte(source), 0644); writeError != nil {
		test.Fatal(writeError)
	}
	engine, prepareError := prepare(test, root, test.TempDir())
	if prepareError != nil {
		test.Fatal(prepareError)
	}
	report, runFilterError := engine.runFilter("large.js", 0, false)
	if runFilterError != nil {
		test.Fatal(runFilterError)
	}
	if report.Pass != 1 {
		test.Fatalf("large program: %+v", report)
	}
	generated, readError := os.ReadFile(
		filepath.Join(engine.programDirectory(classify("large.js", source, false)), "program.c"),
	)
	if readError != nil {
		test.Fatal(readError)
	}
	if len(generated) <= outputLimit || generated[len(generated)-1] != '\n' {
		test.Fatalf("fixture must exceed old limit and end with newline: %d bytes", len(generated))
	}
}

func TestOutputOverflowIsReported(test *testing.T) {
	test.Parallel()
	buffer := limitedBuffer{limit: 3}
	written, writeError := buffer.Write([]byte("abcdef"))
	if writeError != nil || written != 6 || !buffer.exceeded || buffer.String() != "abc" {
		test.Fatalf("overflow: written=%d error=%v buffer=%+v", written, writeError, buffer)
	}
}
