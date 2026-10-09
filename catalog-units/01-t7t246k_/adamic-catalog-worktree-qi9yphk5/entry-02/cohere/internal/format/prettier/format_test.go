package prettier

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/format/formatoptions"
)

// TestMain clears ForkPathVariable once for the package, so every engine a test builds loads the embedded
// bundles whatever the developer's shell exports. Clearing it per test with t.Setenv would hold every test
// that builds an engine serial, since a test that calls t.Setenv cannot run in parallel. The tests in
// forkpath_test.go that set the variable do so with t.Setenv, serially, and go test finishes every serial
// test, restoring what it set, before any parallel test runs.
func TestMain(m *testing.M) {
	if err := os.Unsetenv(ForkPathVariable); err != nil {
		fmt.Fprintf(os.Stderr, "clearing %s: %v\n", ForkPathVariable, err)
		os.Exit(1)
	}
	os.Exit(m.Run())
}

// newTestEngine builds an engine, and no longer skips when a fork is absent.
//
// This function used to stat a directory and skip the whole file when it was missing, because the
// bundles were build output of a checkout that only existed on one machine. That skip is gone, and
// the reason it is gone is the point of the vendoring: the bundles are embedded, so there is no
// machine where they can be absent, and a failure here is a real failure rather than a missing
// dependency.
//
// Worth keeping the record of why that skip was dangerous, because the same shape will be proposed
// again for something else. A skip is truthful only to a reader who reads it, and the reader that
// mattered was a CI summary printing `ok internal/prettier` and swallowing the line. Every assertion
// in this file skipped, and the package reported as passing over an engine that could not load a
// single bundle:
//
//	COHERE_PRETTIER_FORK=/nonexistent go test ./internal/prettier/
//	ok  github.com/system-inc/cohere/internal/prettier  0.208s   exit=0
//
// The message that skip carried reached a `-v` reader and nobody else, and that was not a property
// of `t.Skipf` that some other mechanism dodges. Measured both directions: a `TestMain` writing the
// same warning straight to `os.Stderr` is swallowed in a passing package, and shown in a failing
// one. `go test` suppresses a passing package's output regardless of source, so the only thing that
// changes a plain run is the package not passing -- which forecloses every message-shaped fix and
// left vendoring as the only real one.
//
// The engine loads the embedded bundles because TestMain clears ForkPathVariable for the whole package,
// so a fork a developer's shell points at cannot stand in for them.
func newTestEngine(t *testing.T) *Engine {
	t.Helper()
	engine, err := New(formatoptions.Default())
	if err != nil {
		t.Fatalf("building the engine from the embedded bundles: %v", err)
	}
	return engine
}

// TestFormatsCrampedSource is the known-dirty control. A formatter that has never changed anything
// has not been shown to run at all.
func TestFormatsCrampedSource(t *testing.T) {
	t.Parallel()

	engine := newTestEngine(t)
	formatted, err := engine.Format("probe.ts", "const   x=1\n")
	if err != nil {
		t.Fatalf("format: %v", err)
	}
	if formatted != "const x = 1;\n" {
		t.Fatalf("formatted %q, want %q", formatted, "const x = 1;\n")
	}
}

// TestLeavesFormattedSourceAlone is the other half of the control. A formatter that rewrites
// everything is as useless as one that rewrites nothing.
func TestLeavesFormattedSourceAlone(t *testing.T) {
	t.Parallel()

	engine := newTestEngine(t)
	source := "const x = 1;\n"
	formatted, err := engine.Format("clean.ts", source)
	if err != nil {
		t.Fatalf("format: %v", err)
	}
	if formatted != source {
		t.Fatalf("formatted %q, want it unchanged", formatted)
	}
}

// TestPromiseIsDrained is the regression for the least obvious failure in this package.
//
// prettier.format returns a promise and goja has no event loop, so a version that stringifies the
// result without pumping the job queue writes the literal text "[object Promise]" into the file.
// That is what the first working version produced.
func TestPromiseIsDrained(t *testing.T) {
	t.Parallel()

	engine := newTestEngine(t)
	formatted, err := engine.Format("probe.ts", "const x=1\n")
	if err != nil {
		t.Fatalf("format: %v", err)
	}
	if strings.Contains(formatted, "Promise") {
		t.Fatalf("the promise was not drained: %q", formatted)
	}
}

// TestForkCustomizationsSurvive pins the four divergences our fork carries on top of upstream. They
// are the exact thing a reimplementation loses silently, and the reason we run the fork rather than
// stock Prettier.
func TestForkCustomizationsSurvive(t *testing.T) {
	t.Parallel()

	engine := newTestEngine(t)
	source := "function probe(a) {\n" +
		"    if(a) {\n        a = 1;\n    }\n" +
		"    else {\n        a = 2;\n    }\n" +
		"    for(const x of [a]) {\n        a = x;\n    }\n" +
		"    while(a) {\n        a = 0;\n    }\n" +
		"    switch(a) {\n        default:\n            break;\n    }\n" +
		"    const f = function() {\n        return a;\n    };\n" +
		"    return f();\n}\n"
	formatted, err := engine.Format("probe.js", source)
	if err != nil {
		t.Fatalf("format: %v", err)
	}
	for _, want := range []string{"if(a)", "for(const", "while(a)", "switch(a)", "function()", "\n    else {"} {
		if !strings.Contains(formatted, want) {
			t.Errorf("fork customization lost: %q missing from\n%s", want, formatted)
		}
	}
}

// TestMarkdownTildeIsNotDoubled is the regression for the goja \p{...} bug.
//
// goja implements no Unicode property escapes and returns false for every character rather than
// throwing, which inverted CommonMark's flanking calculation and turned "approximately 25K" into
// strikethrough. The fix lives in the fork's build; this proves the bundles being loaded carry it.
func TestMarkdownTildeIsNotDoubled(t *testing.T) {
	t.Parallel()

	engine := newTestEngine(t)
	source := "Sizes: `Seed` (~25K tokens), `Warm` (~75K tokens).\n"
	formatted, err := engine.Format("probe.md", source)
	if err != nil {
		t.Fatalf("format: %v", err)
	}
	if strings.Contains(formatted, "~~") {
		t.Fatalf("a standalone tilde was doubled, so the bundles predate the fix: %q", formatted)
	}
}

// TestUnhandledTypeIsAnError proves the engine refuses rather than silently passing a file through.
func TestUnhandledTypeIsAnError(t *testing.T) {
	t.Parallel()

	engine := newTestEngine(t)
	if _, err := engine.Format("probe.rb", "puts 1\n"); err == nil {
		t.Fatal("formatting a .rb file succeeded, want an error")
	}
	if engine.Handles("probe.rb") {
		t.Fatal("Handles said yes to .rb")
	}
}

// TestSyntaxErrorSurfaces confirms a malformed file produces an error whose text the pipeline's
// matcher recognizes as a parse failure rather than a formatter crash.
func TestSyntaxErrorSurfaces(t *testing.T) {
	t.Parallel()

	engine := newTestEngine(t)
	_, err := engine.Format("broken.ts", "const x = {{{ broken\n")
	if err == nil {
		t.Fatal("formatting malformed source succeeded, want an error")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "syntaxerror") {
		t.Fatalf("error %q does not name a syntax error, so the pipeline will count it as a crash", err)
	}
}

// TestFilepathReachesPrettier pins that the engine passes the file name, because Prettier's output
// depends on it: `<T,>` on a single-parameter generic arrow is required in .tsx and dropped in .ts, and
// Prettier tells the two apart by the path, not by the parser. Without a filepath both keep the comma.
func TestFilepathReachesPrettier(t *testing.T) {
	t.Parallel()

	engine := newTestEngine(t)
	source := "const identity = <T,>(value: T) => value;\n"
	for fileName, want := range map[string]string{
		"probe.ts":  "const identity = <T>(value: T) => value;\n",
		"probe.tsx": "const identity = <T,>(value: T) => value;\n",
	} {
		formatted, err := engine.Format(fileName, source)
		if err != nil {
			t.Fatalf("%s: %v", fileName, err)
		}
		if formatted != want {
			t.Errorf("%s formatted %q, want %q", fileName, formatted, want)
		}
	}
}
