// Package conformance is the spec's proof: one program for each sentence of docs/0.1.md that says what Adamic
// refuses or accepts, with the outcome that sentence promises. When someone asks whether Adamic allows something,
// the answer is a case here, not a memory. Language (@system_adamic_language) keeps it, one case per ruling.
//
// A case is a program under testdata/<area>/, an .a or a .ts file, whose first lines say what it pins:
//
//	// Spec: <a phrase docs/0.1.md says, verbatim>
//	// Expect: refused <a phrase the refusal must say>
//	// Expect: lowers
//
// A refused case must be refused (lower.Refused, a promise about the language), never only not yet lowered
// (lower.NotYet, a promise about stage 0), and the refusal must say the phrase. A case that lowers is the sound
// sibling of a refusal: the fix the refusal names, held to compile, so a refusal that grew to swallow its own fix
// goes red too. A case whose spec phrase docs/0.1.md no longer says goes red, so the spec and the suite move together.
package conformance

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
)

// expectation is what a case's header promises.
type expectation struct {
	spec    string
	refused bool
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
			found.refused, found.says, expected = true, strings.TrimPrefix(comment, "Expect: refused "), true
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
	if !promise.refused {
		if lowered != nil {
			return fmt.Sprintf("the spec promises this lowers, and stage 0 says: %v", lowered)
		}
		return ""
	}
	var refused *lower.Refused
	if !errors.As(lowered, &refused) {
		return fmt.Sprintf("the spec promises a refusal saying %q, and stage 0 gave %v", promise.says, lowered)
	}
	if !strings.Contains(lowered.Error(), promise.says) {
		return fmt.Sprintf("refused, but without saying %q: %v", promise.says, lowered)
	}
	return ""
}

func TestEveryCaseKeepsTheSpecsPromise(t *testing.T) {
	t.Parallel()
	spec, err := os.ReadFile(filepath.Join("..", "..", "docs", "0.1.md"))
	if err != nil {
		t.Fatal(err)
	}
	var paths []string
	for _, pattern := range []string{"*.a", "*.ts"} {
		matches, err := filepath.Glob(filepath.Join("testdata", "*", pattern))
		if err != nil {
			t.Fatal(err)
		}
		paths = append(paths, matches...)
	}
	if len(paths) == 0 {
		t.Fatal("testdata holds no cases")
	}
	for _, path := range paths {
		t.Run(filepath.ToSlash(strings.TrimPrefix(path, "testdata"+string(filepath.Separator))), func(t *testing.T) {
			t.Parallel()
			promise, err := readExpectation(path)
			if err != nil {
				t.Fatal(err)
			}
			absolute, err := filepath.Abs(path)
			if err != nil {
				t.Fatal(err)
			}
			program, err := load.Load([]string{absolute})
			if err != nil {
				t.Fatalf("Load: %v (a case must be TypeScript tsc accepts; refusing it is Adamic's job)", err)
			}
			_, lowered := lower.Lower(context.Background(), program)
			if broken := judge(promise, string(spec), lowered); broken != "" {
				t.Error(broken)
			}
		})
	}
}

// The judge on promises kept and broken, so a judge that passes everything can't hide behind green cases. Each row
// is a mutant of judge: a NotYet taken for a refusal, a refusal that doesn't say its phrase, any error taken for
// lowering, and a spec phrase the spec doesn't hold.
func TestTheJudgeTellsAKeptPromiseFromABrokenOne(t *testing.T) {
	t.Parallel()
	const spec = "Decorators are refused."
	refusal := &lower.Refused{Where: "main.a:1:1", What: "a decorator", Fix: "write the behavior where it applies"}
	notYet := &lower.NotYet{Where: "main.a:1:1", What: "a decorator"}
	for _, row := range []struct {
		name    string
		promise expectation
		lowered error
		kept    bool
	}{
		{"a refusal saying its phrase", expectation{spec: "Decorators", refused: true, says: "a decorator"}, refusal, true},
		{"a refusal wrapped", expectation{spec: "Decorators", refused: true, says: "a decorator"}, fmt.Errorf("lowering: %w", refusal), true},
		{"not yet is not refused", expectation{spec: "Decorators", refused: true, says: "a decorator"}, notYet, false},
		{"a refusal saying something else", expectation{spec: "Decorators", refused: true, says: "a getter"}, refusal, false},
		{"lowered where a refusal was promised", expectation{spec: "Decorators", refused: true, says: "a decorator"}, nil, false},
		{"lowered as promised", expectation{spec: "Decorators"}, nil, true},
		{"refused where lowering was promised", expectation{spec: "Decorators"}, refusal, false},
		{"not yet where lowering was promised", expectation{spec: "Decorators"}, notYet, false},
		{"a phrase the spec doesn't say", expectation{spec: "Getters", refused: true, says: "a decorator"}, refusal, false},
	} {
		t.Run(row.name, func(t *testing.T) {
			t.Parallel()
			if broken := judge(row.promise, spec, row.lowered); (broken == "") != row.kept {
				t.Errorf("kept %v, judge said %q", row.kept, broken)
			}
		})
	}
}

// A header with no promise is an error, never a case that passes for saying nothing.
func TestAHeaderWithoutAPromiseIsAnError(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	for name, content := range map[string]string{
		"no_spec.a":   "// Expect: lowers\nconsole.log('a');\n",
		"no_expect.a": "// Spec: Decorators\nconsole.log('a');\n",
		"late.a":      "console.log('a');\n// Spec: Decorators\n// Expect: lowers\n",
	} {
		path := filepath.Join(directory, name)
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
		if _, err := readExpectation(path); err == nil {
			t.Errorf("%s: want an error for a header with no promise", name)
		}
	}
}
