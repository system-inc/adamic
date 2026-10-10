package markdowninline

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/nativeproduct"
)

// delimiterExpressionGap is GAPS.md's one gap program, which the gap test runs natively.
const delimiterExpressionGap = "gaps/1_delimiter_expression.ts"

func TestDelimiterExpressionMatchesNode(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(delimiterExpressionGap)
	if err != nil {
		t.Fatal(err)
	}
	answer := onNode(t, path)
	clean(t, "Node gap", answer)
	equal(t, "Node gap", answer.stdout, []byte("a\\*b\n"))
	lowered := gapProgram(t, path)
	native, binary := natively(t, lowered)
	clean(t, "native delimiter expression", native)
	equal(t, "native delimiter expression", native.stdout, answer.stdout)
	backend := onJavaScriptBackend(t, lowered)
	clean(t, "backend delimiter expression", backend)
	equal(t, "backend delimiter expression", backend.stdout, answer.stdout)
	leaks(t, lowered, binary)
}

// gapProgram is the gap at path, loaded and lowered, failing the test if either fails: what the gap test builds
// natively and its TestProduct_ twin builds ahead, from this one function so the two can't drift.
func gapProgram(t *testing.T, path string) *ir.Program {
	t.Helper()
	program, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	lowered, err := lower.Lower(context.Background(), program)
	if err != nil {
		t.Fatal(err)
	}
	return lowered
}

// TestProduct_MarkdownInlineGapNative builds ahead the delimiter expression gap's binaries: the sanitized one natively
// runs, and on macOS the plain one leaks runs.
func TestProduct_MarkdownInlineGapNative(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(delimiterExpressionGap)
	if err != nil {
		t.Fatal(err)
	}
	nativeproduct.Twin(t, gapProgram(t, path))
}
