package tailwind

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
)

// ClassSegmentsIn had no direct test, and a research pass marked it `unverified` for that reason.
//
// It was exercised only through `no-unnecessary-whitespace`, and its ranges were checked once by a
// throwaway probe that no longer exists. That is the gap the marking names: plausibly right, not
// shown right, and the next rule to reach for it would be trusting a probe nobody can re-run.
//
// The property that matters is the one a fix depends on: every segment's range must cover exactly
// its own static text and nothing else. A template head's own Loc spans "`flex  ${" with the
// backtick and the interpolation opener included, so a fix written against the untrimmed range
// deletes both and leaves a file that does not parse.
func TestClassSegmentRangesCoverExactlyTheirText(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name         string
		source       string
		wantSegments []string
	}{
		{
			name:         "plain string is one segment",
			source:       `const element = <div className="flex  items-center" />;`,
			wantSegments: []string{"flex  items-center"},
		},
		{
			name:         "template splits at its hole",
			source:       "const element = <div className={`flex  ${x}  gap-2`} />;",
			wantSegments: []string{"flex  ", "  gap-2"},
		},
		{
			// Three segments, and the middle one has a hole on both sides.
			name:         "template with two holes",
			source:       "const element = <div className={`a ${x} b ${y} c`} />;",
			wantSegments: []string{"a ", " b ", " c"},
		},
		{
			// The trailing segment is empty, which must still carry a valid (empty) range rather
			// than an inverted one.
			name:         "hole at the end",
			source:       "const element = <div className={`px-${size}`} />;",
			wantSegments: []string{"px-", ""},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			segments, source := segmentsOf(t, testCase.source)

			if len(segments) != len(testCase.wantSegments) {
				t.Fatalf("expected %d segments %q, got %d", len(testCase.wantSegments), testCase.wantSegments, len(segments))
			}

			for index, segment := range segments {
				if segment.Text != testCase.wantSegments[index] {
					t.Errorf("segment %d text %q, want %q", index, segment.Text, testCase.wantSegments[index])
				}

				if segment.Range.Pos() > segment.Range.End() {
					t.Fatalf("segment %d has an inverted range [%d,%d), which is a fix that deletes backwards",
						index, segment.Range.Pos(), segment.Range.End())
				}

				covered := source[segment.Range.Pos():segment.Range.End()]
				if covered != segment.Text {
					t.Errorf("segment %d range covers %q but its text is %q, so a fix would rewrite the wrong bytes",
						index, covered, segment.Text)
				}
			}
		})
	}
}

// The hole flags decide whether trimming a segment is safe, so they are asserted directly.
//
// A segment with a hole beside it keeps one whitespace character there, because that space is the
// separator between the substituted value and the neighbouring class. Getting these backwards fuses
// two class names into one that does not exist.
func TestClassSegmentHoleFlags(t *testing.T) {
	t.Parallel()
	segments, _ := segmentsOf(t, "const element = <div className={`a ${x} b ${y} c`} />;")

	if len(segments) != 3 {
		t.Fatalf("expected 3 segments, got %d", len(segments))
	}

	expected := []struct {
		leading  bool
		trailing bool
	}{
		{leading: false, trailing: true},
		{leading: true, trailing: true},
		{leading: true, trailing: false},
	}

	for index, want := range expected {
		if segments[index].LeadingHole != want.leading || segments[index].TrailingHole != want.trailing {
			t.Errorf("segment %d holes are leading=%v trailing=%v, want leading=%v trailing=%v",
				index, segments[index].LeadingHole, segments[index].TrailingHole, want.leading, want.trailing)
		}
	}
}

// A clean template still has segments worth reading.
//
// ClassSegmentsIn deliberately does not route through ClassTemplatesIn, which answers a different
// question and returns nothing when every seam is well spaced. Reusing it here would have skipped
// exactly the templates that are otherwise well written.
func TestClassSegmentsReadTemplatesWithNoBoundaryDefects(t *testing.T) {
	t.Parallel()
	segments, _ := segmentsOf(t, "const element = <div className={`flex  ${x}  gap-2`} />;")

	if len(segments) == 0 {
		t.Fatal("a template whose seams are all clean still has segments, and its doubled spaces " +
			"still need collapsing")
	}
}

func segmentsOf(t *testing.T, source string) ([]ClassSegment, string) {
	t.Helper()

	fileName := tspath.NormalizePath("/Component.tsx")
	sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{
		FileName: tspath.RootedFilePath(fileName), PathKey: tspath.PathKey(fileName),
	}, source, core.ScriptKindTSX)
	if sourceFile == nil {
		t.Fatalf("could not parse %s", source)
	}

	reader := NewClassLiteralReader(DefaultClassLiteralSettings())

	var found []ClassSegment
	var walk func(node *ast.Node)
	walk = func(node *ast.Node) {
		if node == nil {
			return
		}
		found = append(found, reader.ClassSegmentsIn(node)...)
		node.ForEachChild(func(child *ast.Node) bool {
			walk(child)
			return false
		})
	}
	walk(sourceFile.AsNode())

	return found, source
}
