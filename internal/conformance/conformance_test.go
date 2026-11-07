// Package conformance is the spec's proof: one program for each sentence of docs/0.1.md that says what Adamic
// refuses or accepts, with the outcome that sentence promises. When someone asks whether Adamic allows something,
// the answer is a case here, not a memory. Language (@system_adamic_language) keeps it, one case per ruling.
//
// A case is a program under testdata/<area>/, an .a or a .ts file, whose first lines say what it pins:
//
//	// Spec: <a phrase docs/0.1.md says, verbatim>
//	// Expect: refused <a phrase the refusal must say>
//	// Expect: notyet <a phrase the not yet must say>
//	// Expect: lowers
//	// Expect: panics <a phrase the panic must say>
//
// A refused case must be refused (lower.Refused, a promise about the language), never only not yet lowered
// (lower.NotYet, a promise about stage 0), and the refusal must say the phrase. A not yet case is the other promise:
// something the language means to admit, so a refusal there is as wrong as a lowering. A case that lowers is the sound
// sibling of a refusal: the fix the refusal names, held to compile, so a refusal that grew to swallow its own fix
// goes red too. A case whose spec phrase docs/0.1.md no longer says goes red, so the spec and the suite move together.
//
// A panics case pins a Check: an inserted run-time check, where Node runs on or stops otherwise. It must lower, and
// its program, run on the JavaScript backend, must exit 70 with "adamic: panic: <phrase>" on stderr, after printing
// only what the source printed on Node up to there. The JavaScript backend alone is run: native is held to it by the
// oracle's agreement families, and a native build per case would make this package's unit heavy.
package conformance

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/toolchain"
)

// outcome is what lowering a case must end in.
type outcome int

const (
	lowers outcome = iota
	refused
	notYet
	panics
)

// expectation is what a case's header promises.
type expectation struct {
	spec    string
	outcome outcome
	says    string
}

// readExpectation reads a case's header: the comment lines it opens with, up to the first line that isn't one.
func readExpectation(path string) (expectation, error) {
	file, err := os.Open(path)
	if err != nil {
		return expectation{}, err
	}
	defer file.Close()
	var found expectation
	var expected bool
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, "//") {
			break
		}
		comment := strings.TrimSpace(strings.TrimPrefix(line, "//"))
		switch {
		case strings.HasPrefix(comment, "Spec: "):
			found.spec = strings.TrimPrefix(comment, "Spec: ")
		case comment == "Expect: lowers":
			expected = true
		case strings.HasPrefix(comment, "Expect: refused "):
			found.outcome, found.says, expected = refused, strings.TrimPrefix(comment, "Expect: refused "), true
		case strings.HasPrefix(comment, "Expect: notyet "):
			found.outcome, found.says, expected = notYet, strings.TrimPrefix(comment, "Expect: notyet "), true
		case strings.HasPrefix(comment, "Expect: panics "):
			found.outcome, found.says, expected = panics, strings.TrimPrefix(comment, "Expect: panics "), true
		}
	}
	if err := scanner.Err(); err != nil {
		return expectation{}, err
	}
	if found.spec == "" || !expected {
		return expectation{}, fmt.Errorf("%s opens with no // Spec: and // Expect: lines", path)
	}
	return found, nil
}

// judge says how a case broke its promise, given the spec's text and what lowering it returned, or "" when it kept it.
func judge(promise expectation, spec string, lowered error) string {
	if !strings.Contains(spec, promise.spec) {
		return fmt.Sprintf("cites %q, which docs/0.1.md doesn't say", promise.spec)
	}
	switch promise.outcome {
	case refused:
		var refusal *lower.Refused
		if !errors.As(lowered, &refusal) {
			return fmt.Sprintf("the spec promises a refusal saying %q, and stage 0 gave %v", promise.says, lowered)
		}
	case notYet:
		var later *lower.NotYet
		if !errors.As(lowered, &later) {
			return fmt.Sprintf("the spec promises a not yet saying %q, and stage 0 gave %v", promise.says, lowered)
		}
	case panics:
		if lowered != nil {
			return fmt.Sprintf(
				"the spec promises a panic saying %q at run time, and stage 0 didn't lower it: %v",
				promise.says,
				lowered,
			)
		}
		return ""
	default:
		if lowered != nil {
			return fmt.Sprintf("the spec promises this lowers, and stage 0 says: %v", lowered)
		}
		return ""
	}
	if !strings.Contains(lowered.Error(), promise.says) {
		return fmt.Sprintf("stage 0 kept the promise's kind, but without saying %q: %v", promise.says, lowered)
	}
	return ""
}

// run is what a program printed and how it ended.
type run struct {
	stdout, stderr string
	exit           int
}

// judgeRun says how a panics case's run broke its promise, given its JavaScript backend run and the source's run on
// Node, or "" when it kept it: exit 70, the panic saying the phrase, and nothing printed that Node didn't print first.
func judgeRun(promise expectation, backend, source run) string {
	if backend.exit != 70 {
		return fmt.Sprintf(
			"the spec promises a panic saying %q, and the JavaScript backend exited %d (stderr %q)",
			promise.says,
			backend.exit,
			backend.stderr,
		)
	}
	if !strings.Contains(backend.stderr, "adamic: panic: "+promise.says) {
		return fmt.Sprintf("the JavaScript backend panicked without saying %q: %q", promise.says, backend.stderr)
	}
	if !strings.HasPrefix(source.stdout, backend.stdout) {
		return fmt.Sprintf(
			"the JavaScript backend printed %q before its panic, which isn't what Node printed first: %q",
			backend.stdout,
			source.stdout,
		)
	}
	return ""
}

// runOnNode runs path through oracle/node.mjs, as the oracle runs a source or a JavaScript backend's output.
func runOnNode(test *testing.T, path string) run {
	test.Helper()
	runner, err := filepath.Abs(filepath.Join("..", "..", "oracle", "node.mjs"))
	if err != nil {
		test.Fatal(err)
	}
	node, err := toolchain.Path("node")
	if err != nil {
		test.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, node, "--disable-warning=ExperimentalWarning", runner, path)
	command.Env = toolchain.Env()
	var stdout, stderr bytes.Buffer
	command.Stdout, command.Stderr = &stdout, &stderr
	err = command.Run()
	if ctx.Err() != nil {
		test.Fatalf("Node %s: %v", path, ctx.Err())
	}
	exit := 0
	if err != nil {
		var failure *exec.ExitError
		if !errors.As(err, &failure) {
			test.Fatalf("Node %s: %v", path, err)
		}
		exit = failure.ExitCode()
	}
	return run{stdout.String(), stderr.String(), exit}
}

func TestEveryCaseKeepsTheSpecsPromise(test *testing.T) {
	test.Parallel()
	spec, err := os.ReadFile(filepath.Join("..", "..", "docs", "0.1.md"))
	if err != nil {
		test.Fatal(err)
	}
	var paths []string
	for _, pattern := range []string{"*.a", "*.ts"} {
		matches, err := filepath.Glob(filepath.Join("testdata", "*", pattern))
		if err != nil {
			test.Fatal(err)
		}
		paths = append(paths, matches...)
	}
	if len(paths) == 0 {
		test.Fatal("testdata holds no cases")
	}
	for _, path := range paths {
		test.Run(filepath.ToSlash(
			strings.TrimPrefix(path, "testdata"+string(filepath.Separator)),
		), func(test *testing.T) {
			test.Parallel()
			if broken := keeps(test, path, string(spec)); broken != "" {
				test.Error(broken)
			}
		})
	}
}

// keeps says how the case at path broke the spec's promise, or "" when it kept it: lowered and judged, and for a panics
// case, run and judged again.
func keeps(test *testing.T, path, spec string) string {
	test.Helper()
	promise, err := readExpectation(path)
	if err != nil {
		test.Fatal(err)
	}
	absolute, err := filepath.Abs(path)
	if err != nil {
		test.Fatal(err)
	}
	program, err := load.Load([]string{absolute})
	if err != nil {
		test.Fatalf("Load: %v (a case must be TypeScript tsc accepts; refusing it is Adamic's job)", err)
	}
	lowered, err := lower.Lower(context.Background(), program)
	if broken := judge(promise, spec, err); broken != "" || promise.outcome != panics {
		return broken
	}
	generated := filepath.Join(test.TempDir(), "generated.mjs")
	if err := os.WriteFile(generated, []byte(javascript.JavaScript(lowered)), 0o644); err != nil {
		test.Fatal(err)
	}
	return judgeRun(promise, runOnNode(test, generated), runOnNode(test, absolute))
}

// A panics case whose program lowers and runs to the end on Node and on the JavaScript backend is broken, so a judge
// that lowers a panics case and never runs it can't pass for one that does.
func TestAPanicsCaseThatRunsCleanIsBroken(test *testing.T) {
	test.Parallel()
	path := filepath.Join(test.TempDir(), "clean.a")
	source := "// Spec: Array writes\n// Expect: panics index\nconst values: number[] = [1, 2, 3];\nvalues[1] = 9;\nconsole.log(values.join(','));\n"
	if err := os.WriteFile(path, []byte(source), 0o644); err != nil {
		test.Fatal(err)
	}
	if broken := keeps(test, path, "Array writes"); broken == "" {
		test.Error("a panics case that runs clean was judged kept")
	}
}

// The judge on promises kept and broken, so a judge that passes everything can't hide behind green cases. Each row
// is a mutant of judge: a NotYet taken for a refusal or a refusal for a NotYet, a promise that doesn't say its phrase,
// any error taken for lowering, and a spec phrase the spec doesn't hold.
func TestTheJudgeTellsAKeptPromiseFromABrokenOne(test *testing.T) {
	test.Parallel()
	const spec = "Decorators are refused."
	refusal := &lower.Refused{Where: "main.a:1:1", What: "a decorator", Fix: "write the behavior where it applies"}
	later := &lower.NotYet{Where: "main.a:1:1", What: "a decorator"}
	for _, row := range []struct {
		name    string
		promise expectation
		lowered error
		kept    bool
	}{
		{
			"a refusal saying its phrase",
			expectation{spec: "Decorators", outcome: refused, says: "a decorator"},
			refusal,
			true,
		},
		{
			"a refusal wrapped",
			expectation{spec: "Decorators", outcome: refused, says: "a decorator"},
			fmt.Errorf("lowering: %w", refusal),
			true,
		},
		{
			"not yet is not refused",
			expectation{spec: "Decorators", outcome: refused, says: "a decorator"},
			later,
			false,
		},
		{
			"a refusal saying something else",
			expectation{spec: "Decorators", outcome: refused, says: "a getter"},
			refusal,
			false,
		},
		{
			"lowered where a refusal was promised",
			expectation{spec: "Decorators", outcome: refused, says: "a decorator"},
			nil,
			false,
		},
		{"lowered as promised", expectation{spec: "Decorators"}, nil, true},
		{"refused where lowering was promised", expectation{spec: "Decorators"}, refusal, false},
		{"not yet where lowering was promised", expectation{spec: "Decorators"}, later, false},
		{
			"a phrase the spec doesn't say",
			expectation{spec: "Getters", outcome: refused, says: "a decorator"},
			refusal,
			false,
		},
		{
			"a not yet saying its phrase",
			expectation{spec: "Decorators", outcome: notYet, says: "a decorator"},
			later,
			true,
		},
		{
			"refused where not yet was promised",
			expectation{spec: "Decorators", outcome: notYet, says: "a decorator"},
			refusal,
			false,
		},
		{
			"lowered where not yet was promised",
			expectation{spec: "Decorators", outcome: notYet, says: "a decorator"},
			nil,
			false,
		},
		{
			"a not yet saying something else",
			expectation{spec: "Decorators", outcome: notYet, says: "a getter"},
			later,
			false,
		},
		{"a panic case lowered", expectation{spec: "Decorators", outcome: panics, says: "index"}, nil, true},
		{"a panic case refused", expectation{spec: "Decorators", outcome: panics, says: "index"}, refusal, false},
		{"a panic case not yet lowered", expectation{spec: "Decorators", outcome: panics, says: "index"}, later, false},
	} {
		test.Run(row.name, func(test *testing.T) {
			test.Parallel()
			if broken := judge(row.promise, spec, row.lowered); (broken == "") != row.kept {
				test.Errorf("kept %v, judge said %q", row.kept, broken)
			}
		})
	}
}

// The run judge on a panics case's runs, kept and broken. Each row is a mutant of judgeRun: a clean exit taken for a
// panic, another stop's exit taken for a panic, a panic saying another phrase, and a backend that printed something
// Node didn't print first.
func TestTheRunJudgeTellsAPanicFromAnythingElse(test *testing.T) {
	test.Parallel()
	promise := expectation{spec: "Array writes", outcome: panics, says: "index -1 is outside an array of length 3"}
	panicked := run{"before\n", "adamic: panic: index -1 is outside an array of length 3\n", 70}
	node := run{"before\nafter\n", "", 0}
	for _, row := range []struct {
		name            string
		backend, source run
		kept            bool
	}{
		{"a panic saying its phrase after Node's first lines", panicked, node, true},
		{"a panic where Node stops too", panicked, run{"before\n", "RangeError\n", 1}, true},
		{"a clean exit", run{"before\nafter\n", "", 0}, node, false},
		{"an uncaught throw's exit", run{"before\n", panicked.stderr, 1}, node, false},
		{
			"a panic saying something else",
			run{"before\n", "adamic: panic: index 9 is outside an array of length 3\n", 70},
			node,
			false,
		},
		{
			"the phrase without adamic's panic",
			run{"before\n", "index -1 is outside an array of length 3\n", 70},
			node,
			false,
		},
		{"printed what Node didn't", run{"other\n", panicked.stderr, 70}, node, false},
	} {
		test.Run(row.name, func(test *testing.T) {
			test.Parallel()
			if broken := judgeRun(promise, row.backend, row.source); (broken == "") != row.kept {
				test.Errorf("kept %v, judgeRun said %q", row.kept, broken)
			}
		})
	}
}

// A header with no promise is an error, never a case that passes for saying nothing.
func TestAHeaderWithoutAPromiseIsAnError(test *testing.T) {
	test.Parallel()
	directory := test.TempDir()
	for name, content := range map[string]string{
		"no_spec.a":   "// Expect: lowers\nconsole.log('a');\n",
		"no_expect.a": "// Spec: Decorators\nconsole.log('a');\n",
		"late.a":      "console.log('a');\n// Spec: Decorators\n// Expect: lowers\n",
	} {
		path := filepath.Join(directory, name)
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			test.Fatal(err)
		}
		if _, err := readExpectation(path); err == nil {
			test.Errorf("%s: want an error for a header with no promise", name)
		}
	}
}
