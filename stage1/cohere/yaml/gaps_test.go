package yaml

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

func TestLexerGaps(t *testing.T) {
	runner, err := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	if err != nil {
		t.Fatal(err)
	}
	for _, gap := range []struct{ file, output, diagnostic string }{
		{"prefixIncrement.ts", "1\n", "a PrefixUnaryExpression on a number"},
		{"assignmentValue.ts", "1\n", "a BinaryExpression with a number and a number"},
		{"emptyAlternative.ts", "1\n", "an array of never"},
		{"dynamicCase.ts", "1\n", "a case that isn't a constant"},
		{"negativeCase.ts", "1\n", "a case that isn't a constant"},
		{"multiplePush.ts", "2\n", "push with other than one value"},
		{"stringFallback.ts", " \n", "a BinaryExpression with a string and a string"},
		{"valueConjunction.ts", "true\n", "a BinaryExpression with a value and a boolean"},
	} {
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
	closedPresenceGap(t, "gaps/stringPresence.ts")
}

func closedPresenceGap(t *testing.T, file string) {
	t.Helper()
	path, err := filepath.Abs(file)
	if err != nil {
		t.Fatal(err)
	}
	runner, err := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	if err != nil {
		t.Fatal(err)
	}
	expected := run(t, "", nil, "node", "--disable-warning=ExperimentalWarning", runner, path)
	if string(expected) != "false\n" {
		t.Fatalf("Node presence result %q", expected)
	}
	program, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	lowered, err := lower.Lower(context.Background(), program)
	if err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(t.TempDir(), "presence")
	if err := native.Build(native.C(lowered), binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	actual := run(t, "", []string{"ASAN_OPTIONS=detect_leaks=1"}, binary)
	emitted := filepath.Join(t.TempDir(), "presence.mjs")
	if err := os.WriteFile(emitted, []byte(javascript.JavaScript(lowered)), 0644); err != nil {
		t.Fatal(err)
	}
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
	closedPresenceGap(t, "gaps/valuePresence.ts")
}
