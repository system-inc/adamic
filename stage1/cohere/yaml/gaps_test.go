package yaml

import (
	"bytes"
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
)

func TestLexerGaps(t *testing.T) {
	t.Parallel()
	runner, err := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	if err != nil {
		t.Fatal(err)
	}
	for _, gap := range []struct{ file, output, diagnostic string }{
		{"multiplePush.ts", "2\n", "push with other than one value"},
	} {
		// Not parallel: a parallel subtest waits for a slot of its own, and the wait counts in its parent's time.
		t.Run(gap.file, func(t *testing.T) {
			path, err := filepath.Abs(filepath.Join("gaps", gap.file))
			if err != nil {
				t.Fatal(err)
			}
			out := run(t, "", nil, "node", "--disable-warning=ExperimentalWarning", runner, path)
			if string(out) != gap.output {
				t.Fatalf("Node got %q, want %q", out, gap.output)
			}
			program, err := load.Load([]string{path})
			if err != nil {
				t.Fatal(err)
			}
			_, err = lower.Lower(context.Background(), program)
			var notYet *lower.NotYet
			if !errors.As(err, &notYet) || !strings.Contains(err.Error(), gap.diagnostic) {
				t.Fatalf("gap changed or closed: %v; update GAPS.md and remove its workaround", err)
			}
			t.Log(err)
		})
	}
}

func TestStructuralPositionRefusal(t *testing.T) {
	t.Parallel()
	runner, err := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	if err != nil {
		t.Fatal(err)
	}
	path, err := filepath.Abs("gaps/structuralPosition.ts")
	if err != nil {
		t.Fatal(err)
	}
	if out := run(t, "", nil, "node", "--disable-warning=ExperimentalWarning", runner, path); string(out) != "1\n" {
		t.Fatalf("Node got %q", out)
	}
	program, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	_, err = lower.Lower(context.Background(), program)
	var refused *lower.Refused
	if !errors.As(err, &refused) || !strings.Contains(err.Error(), "Span[], an array whose elements can reach back") {
		t.Fatalf("refusal changed or closed: %v", err)
	}
	t.Log(err)
}

func TestClosedStringPresenceGap(t *testing.T) {
	t.Parallel()
	closedGap(t, "gaps/stringPresence.ts", "false\n")
}

// TestClosedLexerGaps keeps the lexer gaps compiler/area-next closed: each proving program now lowers, and
// must print what Node prints from native under the sanitizers and from emitted JavaScript. The port's
// workarounds for them still stand; retiring each one is its own change against the corpus.
func TestClosedLexerGaps(t *testing.T) {
	t.Parallel()
	for _, gap := range []struct{ file, output string }{
		{"assignmentValue.ts", "1\n"},
		{"dynamicCase.ts", "1\n"},
		{"negativeCase.ts", "1\n"},
		{"stringFallback.ts", " \n"},
		{"valueConjunction.ts", "true\n"},
		{"prefixIncrement.ts", "1\n"},
		{"emptyAlternative.ts", "1\n"},
	} {
		// Not parallel: one gap at a time keeps this test to one child at a time, its share.
		t.Run(gap.file, func(t *testing.T) {
			closedGap(t, filepath.Join("gaps", gap.file), gap.output)
		})
	}
}

// closedGap compares file's program, read from the closed gaps' build (port_products_test.go), with what Node prints.
func closedGap(t *testing.T, file, output string) {
	t.Helper()
	binary := yamlClosedGapsBuild.binary(t, file)
	emitted := yamlClosedGapsBuild.emitted(t, file)
	path, err := filepath.Abs(file)
	if err != nil {
		t.Fatal(err)
	}
	runner, err := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	if err != nil {
		t.Fatal(err)
	}
	expected := run(t, "", nil, "node", "--disable-warning=ExperimentalWarning", runner, path)
	if string(expected) != output {
		t.Fatalf("Node printed %q, want %q", expected, output)
	}
	actual := run(t, "", []string{"ASAN_OPTIONS=detect_leaks=1"}, binary)
	backend := run(t, "", nil, "node", "--disable-warning=ExperimentalWarning", runner, emitted)
	for _, side := range []struct {
		name   string
		output []byte
	}{{"native ASan/UBSan/LSan", actual}, {"JavaScript backend", backend}} {
		if !bytes.Equal(side.output, expected) {
			t.Fatalf("%s: %q, Node %q", side.name, side.output, expected)
		}
	}
}

func TestClosedValuePresenceGap(t *testing.T) {
	t.Parallel()
	closedGap(t, "gaps/valuePresence.ts", "false\n")
}
