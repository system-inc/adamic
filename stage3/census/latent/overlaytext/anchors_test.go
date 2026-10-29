// Package overlaytext holds make_overlay.py's plain-text edits on internal/lower to main's own files, so a compiler
// change that reshapes a line the overlay rewrites reds at its author's change, not at the next meter run.
package overlaytext

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// overlayAnchor is one string make_overlay.py finds in an internal/lower file before it edits there. once is its
// `.count(...) == 1` assert; without it, the script's `in` assert, `.index(...)` or bare `.replace(...)` needs the
// anchor at least once, since a replace that finds nothing drops its hook without a word.
type overlayAnchor struct {
	file   string
	anchor string
	once   bool
	lines  string
}

// overlayAnchors mirrors make_overlay.py's own anchors on internal/lower, in its order, by its line numbers; an edit
// there is an edit here. refusals.go, statements.go and latent_state.go are left out: refusalrewrite, statementrewrite
// and statecopy each test theirs. The anchors from line 118 on are counted on the script's output, which its earlier
// inserts never add to, so main's own count is the output's.
var overlayAnchors = []overlayAnchor{
	{"lower.go", "\tfiles := program.Files()", false, "19"},
	{"lower.go", "type lowering struct {", false, "22"},
	{"expression.go", "func (l *lowering) callFunction(call *ast.CallExpression, function int) (ir.Expression, error) {", false, "70-72"},
	{"expression.go", "func (l *lowering) functionValue(node *ast.Node, target int) (ir.Expression, error) {", false, "73-75"},
	{"functions.go", "l.signed[index] = signed{this: this, defaults: defaults, patterns: patterns}", false, "78-80"},
	{"functions.go", "func (l *lowering) lowerFunction(index int, declaration *ast.Node, this int) error {", false, "82-86"},
	{"generic.go", "if err := l.lowerFunction(index, declaration, -1); err != nil {", true, "89-90"},
	{"locals.go", "local, isLocal := l.locals[symbol]", true, "99-100"},
	{"locals.go", "func (l *lowering) declareLocal(name *ast.Node) (int, error) {", true, "129, 135"},
	{"expression.go", "func (l *lowering) expression(node *ast.Node) (ir.Expression, error) {", true, "130, 135"},
	{"functions.go", "\tpending := l.signed[index]", true, "148-149"},
}

// overlayAnchorGaps is every make_overlay.py text edit the internal/lower files read reach would miss, in the
// script's order. read takes a file name within internal/lower.
func overlayAnchorGaps(read func(file string) (string, error)) []string {
	gaps := []string{}
	sources := map[string]string{}
	source := func(file string) (string, bool) {
		if text, ok := sources[file]; ok {
			return text, true
		}
		text, err := read(file)
		if err != nil {
			gaps = append(gaps, err.Error())
			return "", false
		}
		sources[file] = text
		return text, true
	}
	for _, overlayAnchor := range overlayAnchors {
		text, ok := source(overlayAnchor.file)
		if !ok {
			continue
		}
		count := strings.Count(text, overlayAnchor.anchor)
		if overlayAnchor.once && count != 1 {
			gaps = append(gaps, fmt.Sprintf("%s %q (make_overlay.py line %s): found %d, want 1", overlayAnchor.file, overlayAnchor.anchor, overlayAnchor.lines, count))
		} else if count == 0 {
			gaps = append(gaps, fmt.Sprintf("%s %q (make_overlay.py line %s): found 0, want at least 1", overlayAnchor.file, overlayAnchor.anchor, overlayAnchor.lines))
		}
	}
	// Line 20 cuts Lower's body from line 19's anchor to the first lowering struct after it.
	if text, ok := source("lower.go"); ok {
		if start := strings.Index(text, "\tfiles := program.Files()"); start >= 0 && !strings.Contains(text[start:], "\n}\n\ntype lowering") {
			gaps = append(gaps, fmt.Sprintf("lower.go %q after %q (make_overlay.py line 20): found 0, want at least 1", "\n}\n\ntype lowering", "\tfiles := program.Files()"))
		}
	}
	// Lines 120-126: a functions.go with no signatureReturn has signature renamed and a signatureReturn appended, so
	// line 135 counts the appended one. Either way one declaration has to be there exactly once.
	if text, ok := source("functions.go"); ok {
		anchor, lines := "func (l *lowering) signatureReturn(index int, declaration *ast.Node, this int, returns *checker.Type) error {", "131, 135"
		if !strings.Contains(text, "func (l *lowering) signatureReturn(") {
			anchor, lines = "func (l *lowering) signature(index int, declaration *ast.Node, this int) error {", "121-122"
		}
		if count := strings.Count(text, anchor); count != 1 {
			gaps = append(gaps, fmt.Sprintf("functions.go %q (make_overlay.py line %s): found %d, want 1", anchor, lines, count))
		}
	}
	return gaps
}

// readMainLower reads a file from main's own internal/lower.
func readMainLower(file string) (string, error) {
	text, err := os.ReadFile(filepath.Join("../../../../internal/lower", file))
	return string(text), err
}

// Main's own internal/lower holds every anchor make_overlay.py edits by text, as many times as it asserts.
func TestMakeOverlayTextEditsHoldOnMainLower(t *testing.T) {
	t.Parallel()
	// Reads: internal/lower/lower.go
	// Reads: internal/lower/expression.go
	// Reads: internal/lower/functions.go
	// Reads: internal/lower/generic.go
	// Reads: internal/lower/locals.go
	if gaps := overlayAnchorGaps(readMainLower); len(gaps) > 0 {
		t.Fatalf("make_overlay.py's text edits miss main's internal/lower:\n%s", strings.Join(gaps, "\n"))
	}
}

// A reshape is named by file, anchor and make_overlay.py line. Mutants on main's own text: functions.go's pending
// read renamed, signature given a second declaration, and the line-20 cut's lowering struct gone.
func TestMakeOverlayTextEditsNameAReshape(t *testing.T) {
	t.Parallel()
	// Reads: internal/lower/lower.go
	// Reads: internal/lower/expression.go
	// Reads: internal/lower/functions.go
	// Reads: internal/lower/generic.go
	// Reads: internal/lower/locals.go
	mutants := []struct {
		name   string
		file   string
		old    string
		new    string
		wanted []string
	}{
		{"pending renamed", "functions.go", "\tpending := l.signed[index]", "\tpending := l.signed[at]",
			[]string{`functions.go "\tpending := l.signed[index]" (make_overlay.py line 148-149): found 0, want 1`}},
		{"signature twice", "functions.go", "func (l *lowering) signature(index int, declaration *ast.Node, this int) error {",
			"func (l *lowering) signature(index int, declaration *ast.Node, this int) error {\n// func (l *lowering) signature(index int, declaration *ast.Node, this int) error {",
			[]string{`functions.go "func (l *lowering) signature(index int, declaration *ast.Node, this int) error {" (make_overlay.py line 121-122): found 2, want 1`}},
		{"lowering struct gone", "lower.go", "\ntype lowering struct {", "\ntype lowered struct {",
			[]string{
				`lower.go "type lowering struct {" (make_overlay.py line 22): found 0, want at least 1`,
				`lower.go "\n}\n\ntype lowering" after "\tfiles := program.Files()" (make_overlay.py line 20): found 0, want at least 1`,
			}},
	}
	for _, mutant := range mutants {
		gaps := overlayAnchorGaps(func(file string) (string, error) {
			text, err := readMainLower(file)
			if err != nil || file != mutant.file {
				return text, err
			}
			if !strings.Contains(text, mutant.old) {
				return "", fmt.Errorf("mutant %s: main's %s no longer holds %q", mutant.name, file, mutant.old)
			}
			return strings.Replace(text, mutant.old, mutant.new, 1), nil
		})
		if !slices.Equal(gaps, mutant.wanted) {
			t.Errorf("mutant %s: gaps = %q, want %q", mutant.name, gaps, mutant.wanted)
		}
	}
}
