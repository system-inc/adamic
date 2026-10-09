package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestClassifyCorpus locks the skip rules, including negative parse and early error. A runner that
// ignores a negative expectation classifies those fixtures as attempted, and this fails.
func TestClassifyCorpus(t *testing.T) {
	t.Parallel()
	matches, err := filepath.Glob("testdata/corpus/*.js")
	if err != nil {
		t.Fatal(err)
	}
	if len(matches) == 0 {
		t.Fatal("no corpus fixtures")
	}
	for _, path := range matches {
		t.Run(filepath.Base(path), func(t *testing.T) {
			t.Parallel()
			source, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			want, err := os.ReadFile(strings.TrimSuffix(path, ".js") + ".want")
			if err != nil {
				t.Fatal(err)
			}
			lines := strings.Split(strings.TrimSpace(string(want)), "\n")
			got := classify(filepath.Base(path), string(source), false)
			kind := "attempted"
			if got.Skip != "" {
				kind = "skipped"
			}
			if kind != lines[0] {
				t.Errorf("%s: kind %s, want %s (skip %q)", path, kind, lines[0], got.Skip)
			}
			if kind == "skipped" && !strings.Contains(got.Skip, lines[1]) {
				t.Errorf("%s: skip %q, want it to contain %q", path, got.Skip, lines[1])
			}
			if kind == "attempted" && len(lines) > 1 {
				gotPhase := got.NegativePhase + " " + got.NegativeType
				if strings.TrimSpace(gotPhase) != lines[1] {
					t.Errorf("%s: negative %q, want %q", path, gotPhase, lines[1])
				}
			}
			if kind == "attempted" && !strings.Contains(got.Program, "function assertSameValue") {
				t.Errorf("%s: attempted program has no prelude", path)
			}
		})
	}
}

// TestVerdictCorpus locks the comparison. A runner that counts a native failure as a pass fails
// native_fails.json. A runner that ignores a negative expectation fails negative_observed.json,
// because both sides exited non-zero and an ordinary comparison only passes when both exit 0.
func TestVerdictCorpus(t *testing.T) {
	t.Parallel()
	matches, err := filepath.Glob("testdata/verdicts/*.json")
	if err != nil {
		t.Fatal(err)
	}
	if len(matches) == 0 {
		t.Fatal("no verdict fixtures")
	}
	for _, path := range matches {
		t.Run(filepath.Base(path), func(t *testing.T) {
			t.Parallel()
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			var fixture struct {
				verdictInput
				Name string `json:"name"`
				Want string `json:"want"`
			}
			if err := json.Unmarshal(data, &fixture); err != nil {
				t.Fatal(err)
			}
			got := decide(fixture.verdictInput)
			if string(got.Kind) != fixture.Want {
				t.Errorf("%s (%s): got %s (%s), want %s", path, fixture.Name, got.Kind, got.Reason, fixture.Want)
			}
		})
	}
}

func TestNormalizeReason(t *testing.T) {
	t.Parallel()
	cases := []struct {
		stderr string
		want   string
	}{
		{"adamic: prog.ts:1:14: stage 0 can't lower a rest parameter yet\n", "not yet: a rest parameter"},
		{"adamic: /tmp/work/program.ts:1:1: Adamic 0.1 refuses var; use const or let\n", "refuses var"},
		{"program.ts:1:12: error TS7006: Parameter 'x' implicitly has an 'any' type.\n", "error TS7006: Parameter '…' implicitly has an '…' type."},
	}
	for _, test := range cases {
		t.Run(test.stderr, func(t *testing.T) {
			t.Parallel()
			kind, reason := compileClass(test.stderr, 1, false)
			if kind != "refused" {
				t.Errorf("%q: kind %s, want refused", test.stderr, kind)
			}
			if reason != test.want {
				t.Errorf("normalize %q:\n got %q\nwant %q", test.stderr, reason, test.want)
			}
		})
	}
	kind, _ := compileClass("panic: something\ngoroutine 1 [running]:\n", 2, false)
	if kind != "crashed" {
		t.Errorf("compiler panic: kind %s, want crashed", kind)
	}
}

func TestRewriteHarnessCalls(t *testing.T) {
	t.Parallel()
	source := "const text = \"assert.sameValue(1, 2)\";\nassert.sameValue(1, 1);\nassert.throws(TypeError, function () {});\nassert.notSameValue(0, -0);\n"
	got := rewriteHarnessCalls(source)
	if strings.Contains(got, "\"assertSameValue") {
		t.Fatalf("rewrote inside a string:\n%s", got)
	}
	if !strings.Contains(got, "assertSameValue(1, 1);") {
		t.Fatalf("missed sameValue:\n%s", got)
	}
	if !strings.Contains(got, "assertThrows(\"TypeError\", function () {});") {
		t.Fatalf("missed throws:\n%s", got)
	}
	if !strings.Contains(got, "assertNotSameValue(0, -0);") {
		t.Fatalf("missed notSameValue:\n%s", got)
	}
	template := "assert.sameValue(`${'assert.sameValue'}`, 'x');\n"
	rewritten := rewriteHarnessCalls(template)
	if rewritten != "assertSameValue(`${'assert.sameValue'}`, 'x');\n" {
		t.Fatalf("template: %s", rewritten)
	}
	nested := "assert.sameValue(`${assert.sameValue(1, 1)}`, 'x');\n"
	rewritten = rewriteHarnessCalls(nested)
	if !strings.Contains(rewritten, "assertSameValue(`${assertSameValue(1, 1)}`, 'x');") {
		t.Fatalf("nested call: %s", rewritten)
	}
}

func TestFrontmatterShapes(t *testing.T) {
	t.Parallel()
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
		t.Fatal("frontmatter not found")
	}
	if strings.Join(got.Features, ",") != "String.prototype.padStart,arrow-function" {
		t.Fatalf("features %v", got.Features)
	}
	if strings.Join(got.Flags, ",") != "onlyStrict,generated" {
		t.Fatalf("flags %v", got.Flags)
	}
	if strings.Join(got.Includes, ",") != "compareArray.js" {
		t.Fatalf("includes %v", got.Includes)
	}
	if got.NegativePhase != "runtime" || got.NegativeType != "TypeError" {
		t.Fatalf("negative %s %s", got.NegativePhase, got.NegativeType)
	}
	classified := classify("built-ins/String/x.js", source, false)
	if classified.Skip != "" {
		t.Fatalf("a padStart test was skipped: %s", classified.Skip)
	}
}

// TestMiniRunner runs the real command path over a three-file checkout: one test stage 0 and Node
// both pass, one negative parse test that must be skipped, and one `var` test the compiler refuses.
// Counting that refusal as a pass fails this.
// Not parallel: writes shared $UserCacheDir/adamic/runtime and $UserCacheDir/adamic/test262 caches.
func TestMiniRunner(t *testing.T) {
	if _, err := exec.LookPath("clang"); err != nil {
		t.Skip("clang not on PATH")
	}
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node not on PATH")
	}
	work := t.TempDir()
	engine, err := prepare("../..", "testdata/mini", work)
	if err != nil {
		t.Fatal(err)
	}
	passed, err := engine.runFilter("pass", 0, false)
	if err != nil {
		t.Fatal(err)
	}
	if passed.Pass != 1 || passed.Fail != 0 || passed.Refused != 0 || passed.Crashed != 0 {
		t.Fatalf("pass fixture: %+v reasons fail=%v refused=%v crash=%v", passed.Pass, passed.FailReasons, passed.RefusalReasons, passed.CrashReasons)
	}
	skipped, err := engine.runFilter("skip", 0, false)
	if err != nil {
		t.Fatal(err)
	}
	if skipped.Skipped != 1 || skipped.Pass != 0 {
		t.Fatalf("negative fixture: skipped %d pass %d reasons %v", skipped.Skipped, skipped.Pass, skipped.SkipReasons)
	}
	refused, err := engine.runFilter("refuse", 0, false)
	if err != nil {
		t.Fatal(err)
	}
	if refused.Refused != 1 || refused.Pass != 0 {
		t.Fatalf("var fixture: refused %d pass %d reasons %v", refused.Refused, refused.Pass, refused.RefusalReasons)
	}
	if len(refused.RefusalReasons) == 0 || !strings.Contains(refused.RefusalReasons[0].Reason, "refuses var") {
		t.Fatalf("var fixture reason: %+v", refused.RefusalReasons)
	}
}

// A large sort previously reached clang with its C cut at 256 KiB, falsely blaming the emitter.
// Not parallel: writes shared $UserCacheDir/adamic/runtime and $UserCacheDir/adamic/test262 caches.
func TestLargeCompilerOutputIsComplete(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "test"), 0755); err != nil {
		t.Fatal(err)
	}
	source := strings.Repeat("assert.sameValue(1, 1);\n", 1800)
	if err := os.WriteFile(filepath.Join(root, "test", "large.js"), []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
	engine, err := prepare("../..", root, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	report, err := engine.runFilter("large.js", 0, false)
	if err != nil {
		t.Fatal(err)
	}
	if report.Pass != 1 {
		t.Fatalf("large program: %+v", report)
	}
	generated, err := os.ReadFile(filepath.Join(engine.programDirectory(classify("large.js", source, false)), "program.c"))
	if err != nil {
		t.Fatal(err)
	}
	if len(generated) <= outputLimit || generated[len(generated)-1] != '\n' {
		t.Fatalf("fixture must exceed old limit and end with newline: %d bytes", len(generated))
	}
}

func TestOutputOverflowIsReported(t *testing.T) {
	t.Parallel()
	buffer := limitedBuffer{limit: 3}
	written, err := buffer.Write([]byte("abcdef"))
	if err != nil || written != 6 || !buffer.exceeded || buffer.String() != "abc" {
		t.Fatalf("overflow: written=%d error=%v buffer=%+v", written, err, buffer)
	}
}
