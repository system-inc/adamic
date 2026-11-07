package parser

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/childguard"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/leakcheck"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

// Not parallel: ASAN_OPTIONS and UBSAN_OPTIONS process environment via test.Setenv.
func TestClosedStrongAstParentGap(test *testing.T) {
	path, absolutePathError := filepath.Abs("gaps/1_strong_ast_parent.ts")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	runner, absolutePathError := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	expected := execute(test, "", "node", "--disable-warning=ExperimentalWarning", runner, path).output
	if string(expected) != "root\n" {
		test.Fatalf("Node gap result %q", expected)
	}
	program, loadError := load.Load([]string{path})
	if loadError != nil {
		test.Fatal(loadError)
	}
	lowered, lowerError := lower.Lower(context.Background(), program)
	if lowerError != nil {
		test.Fatal(lowerError)
	}
	if len(lowered.GraphTypes) == 0 {
		test.Fatal("strong parent/child cycle accepted without graph ownership")
	}
	code := native.C(lowered)
	binary := gapNative(test, "gaps/1_strong_ast_parent.ts", "", native.Options{Sanitize: true})
	test.Setenv("ASAN_OPTIONS", "detect_leaks=1")
	test.Setenv("UBSAN_OPTIONS", "halt_on_error=1")
	actual := execute(test, "", binary).output
	emitted := filepath.Join(test.TempDir(), "strong-parent.mjs")
	if writeError := os.WriteFile(emitted, []byte(javascript.JavaScript(lowered)), 0644); writeError != nil {
		test.Fatal(writeError)
	}
	backend := execute(test, "", "node", "--disable-warning=ExperimentalWarning", runner, emitted).output
	for _, side := range []struct {
		name   string
		output []byte
	}{{"native ASan/UBSan/LSan", actual}, {"JavaScript backend", backend}} {
		if !bytes.Equal(side.output, expected) {
			test.Fatalf("%s: %q, Node %q", side.name, side.output, expected)
		}
	}
	if report := leakcheck.Report(test, code, binary); report != "" {
		test.Fatal(report)
	}
	counted := gapNative(test, "gaps/1_strong_ast_parent.ts", "", native.Options{Count: true})
	timeoutContext, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	stdout, stderr, captureError := childguard.Capture(timeoutContext, exec.Command(counted))
	if captureError != nil || !bytes.Equal(stdout, expected) {
		test.Fatalf("counted native: %v, stdout %q, stderr %q; Node %q", captureError, stdout, stderr, expected)
	}
	if report := leakcheck.Unbalanced(leakcheck.Run{Stdout: stdout, Stderr: stderr}); report != "" {
		test.Fatal(report)
	}
	test.Logf("Node, JavaScript, and sanitized native agree on %q; leak check clean; %s", expected, stderr)
}

// Not parallel: ASAN_OPTIONS process environment via test.Setenv.
func TestPushSpreadGap(test *testing.T) {
	path, absolutePathError := filepath.Abs("gaps/2_push_spread.ts")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	runner, absolutePathError := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	expected := execute(test, "", "node", "--disable-warning=ExperimentalWarning", runner, path).output
	if string(expected) != "1,2,3\n" {
		test.Fatalf("Node gap result %q", expected)
	}
	program, loadError := load.Load([]string{path})
	if loadError != nil {
		test.Fatal(loadError)
	}
	lowered, lowerError := lower.Lower(context.Background(), program)
	if lowerError != nil {
		test.Fatal(lowerError)
	}
	binary := gapNative(test, "gaps/2_push_spread.ts", "", native.Options{Sanitize: true})
	test.Setenv("ASAN_OPTIONS", "detect_leaks=1")
	actual := execute(test, "", binary).output
	emitted := filepath.Join(test.TempDir(), "spread.mjs")
	if writeError := os.WriteFile(emitted, []byte(javascript.JavaScript(lowered)), 0644); writeError != nil {
		test.Fatal(writeError)
	}
	backend := execute(test, "", "node", "--disable-warning=ExperimentalWarning", runner, emitted).output
	for _, side := range []struct {
		name   string
		output []byte
	}{{"native ASan/UBSan/LSan", actual}, {"JavaScript backend", backend}} {
		if !bytes.Equal(side.output, expected) {
			test.Fatalf("%s: %q, Node %q", side.name, side.output, expected)
		}
	}
}

func TestClassMethodInterfaceGap(test *testing.T) {
	test.Parallel()
	path, absolutePathError := filepath.Abs("gaps/4_class_interface_method.ts")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	runner, absolutePathError := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	result := execute(test, "", "node", "--disable-warning=ExperimentalWarning", runner, path)
	if string(result.output) != "1\n" {
		test.Fatalf("Node gap result %q", result.output)
	}
	program, loadError := load.Load([]string{path})
	if loadError != nil {
		test.Fatal(loadError)
	}
	_, lowerError := lower.Lower(context.Background(), program)
	var notYet *lower.NotYet
	if !errors.As(lowerError, &notYet) ||
		notYet.What != "a class method through a view that erases its prototype origin" {
		test.Fatalf("GAPS.md records an erased class-method origin NotYet, got %v", lowerError)
	}
	test.Logf("Node prints 1; Adamic reports: %v", notYet)
}

// Not parallel: ASAN_OPTIONS process environment via test.Setenv.
func TestClosedOptionalFunctionValueGap(test *testing.T) {
	path, absolutePathError := filepath.Abs("gaps/5_optional_function_value.ts")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	runner, absolutePathError := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	expected := execute(test, "", "node", "--disable-warning=ExperimentalWarning", runner, path).output
	if string(expected) != "1\n" {
		test.Fatalf("Node optional function result %q", expected)
	}
	program, loadError := load.Load([]string{path})
	if loadError != nil {
		test.Fatal(loadError)
	}
	lowered, lowerError := lower.Lower(context.Background(), program)
	if lowerError != nil {
		test.Fatal(lowerError)
	}
	binary := gapNative(test, "gaps/5_optional_function_value.ts", "", native.Options{Sanitize: true})
	test.Setenv("ASAN_OPTIONS", "detect_leaks=1")
	actual := execute(test, "", binary).output
	emitted := filepath.Join(test.TempDir(), "optional.mjs")
	if writeError := os.WriteFile(emitted, []byte(javascript.JavaScript(lowered)), 0644); writeError != nil {
		test.Fatal(writeError)
	}
	backend := execute(test, "", "node", "--disable-warning=ExperimentalWarning", runner, emitted).output
	for _, side := range []struct {
		name   string
		output []byte
	}{{"native ASan/UBSan/LSan", actual}, {"JavaScript backend", backend}} {
		if !bytes.Equal(side.output, expected) {
			test.Fatalf("%s: %q, Node %q", side.name, side.output, expected)
		}
	}
}

// TestClosedConditionalEmptyArrayGap: the untyped [] branch of a conditional lowers on
// compiler/area-stack (views slice 1, Oct 8), held to Node on native ASan/UBSan/LSan and the
// JavaScript backend. The driver's number[] annotation still stands.
// Not parallel: ASAN_OPTIONS process environment via test.Setenv.
func TestClosedConditionalEmptyArrayGap(test *testing.T) {
	path, absolutePathError := filepath.Abs("gaps/6_conditional_empty_array.ts")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	runner, absolutePathError := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	expected := execute(test, "", "node", "--disable-warning=ExperimentalWarning", runner, path).output
	if string(expected) != "1\n" {
		test.Fatalf("Node gap result %q", expected)
	}
	program, loadError := load.Load([]string{path})
	if loadError != nil {
		test.Fatal(loadError)
	}
	lowered, lowerError := lower.Lower(context.Background(), program)
	if lowerError != nil {
		test.Fatal(lowerError)
	}
	binary := gapNative(test, "gaps/6_conditional_empty_array.ts", "", native.Options{Sanitize: true})
	test.Setenv("ASAN_OPTIONS", "detect_leaks=1")
	actual := execute(test, "", binary).output
	emitted := filepath.Join(test.TempDir(), "conditional.mjs")
	if writeError := os.WriteFile(emitted, []byte(javascript.JavaScript(lowered)), 0644); writeError != nil {
		test.Fatal(writeError)
	}
	backend := execute(test, "", "node", "--disable-warning=ExperimentalWarning", runner, emitted).output
	for _, side := range []struct {
		name   string
		output []byte
	}{{"native ASan/UBSan/LSan", actual}, {"JavaScript backend", backend}} {
		if !bytes.Equal(side.output, expected) {
			test.Fatalf("%s: %q, Node %q", side.name, side.output, expected)
		}
	}
}
