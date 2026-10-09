package tailwind

import (
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// Expectations measured by running `better-tailwindcss/no-unnecessary-whitespace` with `--fix` over
// probe fixtures, so the fix text below is what upstream actually writes rather than what its source
// suggests it writes.
//
//	"flex  items-center"   ->  "flex items-center"
//	" flex"                ->  "flex"
//	"flex "                ->  "flex"
//	"   "                  ->  ""            (emptied, not deleted)
//	"flex items-center"     silent
//	""                      silent
//	`flex  ${x}`           ->  `flex ${x}`   (one space kept at the hole)
//	`flex ${x}  gap-2`     ->  `flex ${x} gap-2`
//	"flex\titems-center"    silent           (a tab separates as well as a space)

func TestNoUnnecessaryWhitespaceReportsPadding(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name     string
		fileName string
		source   string
		wantIds  []string
	}{
		{
			name:     "doubled space between classes",
			fileName: "Component.tsx",
			source:   `const element = <div className="flex  items-center" />;`,
			wantIds:  []string{"unnecessaryWhitespace"},
		},
		{
			name:     "leading space",
			fileName: "Component.tsx",
			source:   `const element = <div className=" flex" />;`,
			wantIds:  []string{"unnecessaryWhitespace"},
		},
		{
			name:     "trailing space",
			fileName: "Component.tsx",
			source:   `const element = <div className="flex " />;`,
			wantIds:  []string{"unnecessaryWhitespace"},
		},
		{
			// Emptied rather than deleted, matching upstream. Whether the attribute should exist is
			// not this rule's opinion.
			name:     "whitespace only",
			fileName: "Component.tsx",
			source:   `const element = <div className="   " />;`,
			wantIds:  []string{"unnecessaryWhitespace"},
		},
		{
			name:     "inside a class-merging call",
			fileName: "Component.tsx",
			source:   `const merged = cn('flex  items-center');`,
			wantIds:  []string{"unnecessaryWhitespace"},
		},
		{
			name:     "assigned to a class-named variable",
			fileName: "Styles.ts",
			source:   `const className = 'px-4  py-2';`,
			wantIds:  []string{"unnecessaryWhitespace"},
		},
		{
			// A template segment before a hole. The doubled space is still wrong; the single space
			// that survives it is not.
			name:     "doubled space before a hole",
			fileName: "Component.tsx",
			source:   "const element = <div className={`flex  ${extra}`} />;",
			wantIds:  []string{"unnecessaryWhitespace"},
		},
		{
			// A template segment after a hole.
			name:     "doubled space after a hole",
			fileName: "Component.tsx",
			source:   "const element = <div className={`${extra}  flex`} />;",
			wantIds:  []string{"unnecessaryWhitespace"},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoUnnecessaryWhitespace, testCase.fileName, testCase.source)
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
		})
	}
}

// The clean half. The template cases here are the ones a naive trim breaks, because the whitespace
// next to a hole looks like padding and is actually the separator.
func TestNoUnnecessaryWhitespaceStaysSilent(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name     string
		fileName string
		source   string
	}{
		{
			name:     "single spaces",
			fileName: "Component.tsx",
			source:   `const element = <div className="flex items-center gap-2" />;`,
		},
		{
			name:     "one class",
			fileName: "Component.tsx",
			source:   `const element = <div className="flex" />;`,
		},
		{
			name:     "empty string",
			fileName: "Component.tsx",
			source:   `const element = <div className="" />;`,
		},
		{
			// A tab separates two class names as well as a space does, and upstream does not report
			// it. Collapsing tabs would rewrite every tab-indented multi-line class list in the tree.
			name:     "tab between classes",
			fileName: "Component.tsx",
			source:   "const element = <div className=\"flex\titems-center\" />;",
		},
		{
			// The whitespace at each hole is exactly one space, which is correct and must survive.
			name:     "single space around a hole",
			fileName: "Component.tsx",
			source:   "const element = <div className={`flex ${extra} gap-2`} />;",
		},
		{
			// No whitespace at the hole at all. That is `no-concatenated-classes`'s finding, and
			// this rule must not invent a space to fix it, because that would change the class names.
			name:     "no whitespace at the hole",
			fileName: "Component.tsx",
			source:   "const element = <div className={`px-${size}`} />;",
		},
		{
			// Adjacent holes with nothing between them.
			name:     "adjacent holes",
			fileName: "Component.tsx",
			source:   "const element = <div className={`${a}${b}`} />;",
		},
		{
			name:     "unrelated attribute",
			fileName: "Component.tsx",
			source:   `const element = <div title="flex  items-center" />;`,
		},
		{
			name:     "unrelated callee",
			fileName: "Component.tsx",
			source:   `const value = someOtherFunction('flex  items-center');`,
		},
		{
			name:     "unrelated variable",
			fileName: "Styles.ts",
			source:   `const description = 'flex  items-center';`,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoUnnecessaryWhitespace, testCase.fileName, testCase.source)
			rule_testing.ExpectClean(t, result)
		})
	}
}

// TestNoUnnecessaryWhitespaceFixMatchesUpstream is the half that matters most here.
//
// This rule rewrites source unattended, and the interesting cases are the template ones where the
// obviously correct repair (trim it) is wrong. Each expectation is what upstream's `--fix` actually
// produced.
func TestNoUnnecessaryWhitespaceFixMatchesUpstream(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name   string
		source string
		want   string
	}{
		{
			name:   "collapses a doubled space",
			source: `const element = <div className="flex  items-center" />;`,
			want:   `const element = <div className="flex items-center" />;`,
		},
		{
			name:   "drops a leading space",
			source: `const element = <div className=" flex" />;`,
			want:   `const element = <div className="flex" />;`,
		},
		{
			name:   "drops a trailing space",
			source: `const element = <div className="flex " />;`,
			want:   `const element = <div className="flex" />;`,
		},
		{
			name:   "empties a whitespace-only string",
			source: `const element = <div className="   " />;`,
			want:   `const element = <div className="" />;`,
		},
		{
			// The load-bearing case: one space survives at the hole.
			name:   "keeps one space before a hole",
			source: "const element = <div className={`flex  ${extra}`} />;",
			want:   "const element = <div className={`flex ${extra}`} />;",
		},
		{
			name:   "keeps one space after a hole",
			source: "const element = <div className={`${extra}  flex`} />;",
			want:   "const element = <div className={`${extra} flex`} />;",
		},
		{
			// Upstream's own case ("should remove newlines whenever possible"): with allowMultiline on,
			// the default, a newline run keeps its newline and indentation and only doubled spaces
			// between classes are shortened.
			name:   "keeps a wrapped list wrapped",
			source: "const element = <img className={`\n      d  c\n      b  a\n    `} />;",
			want:   "const element = <img className={`\n      d c\n      b a\n    `} />;",
		},
		{
			name:   "drops the spaces before a newline",
			source: "const element = <div className={`flex  \n    items-center`} />;",
			want:   "const element = <div className={`flex\n    items-center`} />;",
		},
		{
			// A run longer than one character becomes one space, whatever it was made of.
			name:   "writes a doubled tab as one space",
			source: "const element = <div className=\"flex\t\titems-center\" />;",
			want:   "const element = <div className=\"flex items-center\" />;",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoUnnecessaryWhitespace, "Component.tsx", testCase.source)
			rule_testing.ExpectFixedSource(t, result, testCase.want)
		})
	}
}

// TestWhitespaceFixClaimsNoClassByte pins the property the other fixers depend on.
//
// Every edit this rule proposes deletes whitespace and nothing else, and never the first character
// of a separator between two classes, because those bytes are what `no-duplicate-classes` and
// `enforce-consistent-class-order` edit in the same pass (class_tokens.go). A rewrite of the whole
// literal produces the same text and claims every byte, and the engine then refuses one of the two.
func TestWhitespaceFixClaimsNoClassByte(t *testing.T) {
	t.Parallel()
	source := `const element = <div className="  flex   items-center  gap-2 " />;`
	result := rule_testing.Run(t, NoUnnecessaryWhitespace, "Component.tsx", source)
	if len(result.Diagnostics) != 1 {
		t.Fatalf("expected one finding, got %d", len(result.Diagnostics))
	}
	fixes := result.Diagnostics[0].Fixes
	for _, fix := range fixes {
		removed := source[fix.Range.Pos():fix.Range.End()]
		if fix.Text != "" || strings.TrimSpace(removed) != "" {
			t.Errorf("a fix rewrites %q to %q, which is more than deleting whitespace", removed, fix.Text)
		}
	}

	// The first byte of each separator that has a class on both sides: after `flex` and after
	// `items-center`. Neither may sit inside any fix.
	contentStart := strings.Index(source, `"`) + 1
	contentEnd := strings.LastIndex(source, `"`)
	content := source[contentStart:contentEnd]
	for _, className := range []string{"flex", "items-center"} {
		separatorStart := contentStart + strings.Index(content, className) + len(className)
		for _, fix := range fixes {
			if fix.Range.Pos() <= separatorStart && separatorStart < fix.Range.End() {
				t.Errorf("a fix covers the separator's first byte after %q, which another rule may edit", className)
			}
		}
	}
	if len(fixes) != 4 {
		t.Errorf("expected four deletions (two padding runs, two excesses), got %d", len(fixes))
	}
	rule_testing.ExpectFixedSource(t, result, `const element = <div className="flex items-center gap-2" />;`)
}

// TestWhitespaceFixNeverFusesClassesAcrossAHole is the known-dirty control.
//
// The tempting implementation is `strings.Join(strings.Fields(text), " ")` applied to the whole
// segment, which is correct for a plain string and silently destructive inside a template: it trims
// the space next to the interpolation, so `flex ${size}` becomes `flex${size}` and the two fuse into
// one class name that does not exist. The damage is invisible in review and only shows up as a
// missing style at runtime.
//
// So the naive rewrite is computed here and required to differ from what the rule proposes.
func TestWhitespaceFixNeverFusesClassesAcrossAHole(t *testing.T) {
	t.Parallel()
	// A segment sitting between two holes, with padding that a plain trim would remove entirely.
	segment := ClassSegment{
		Text:         "  flex  ",
		LeadingHole:  true,
		TrailingHole: true,
	}

	tidied, needsTidying := tidyWhitespace(segment, true)
	if !needsTidying {
		t.Fatal("a doubled space inside the segment is a defect and should have been reported")
	}

	naive := "flex"
	if tidied == naive {
		t.Fatal("the fix trimmed the separators next to the interpolations, so the substituted values " +
			"would fuse with the class and produce names that do not exist")
	}
	if tidied != " flex " {
		t.Fatalf("expected one space preserved on each side of the hole, got %q", tidied)
	}

	// And the discrimination has to be real: with no holes, the same text trims fully.
	plain := ClassSegment{Text: "  flex  "}
	plainTidied, plainNeedsTidying := tidyWhitespace(plain, true)
	if !plainNeedsTidying || plainTidied != "flex" {
		t.Fatalf("outside a template the padding should be removed entirely, got %q", plainTidied)
	}
}

// TestNoUnnecessaryWhitespaceAllowMultiline is upstream's `allowMultiline`, each way, on upstream's own
// case. On, the default, a class list written across lines stays across lines. Off, it is whitespace
// like any other and the list becomes one line.
func TestNoUnnecessaryWhitespaceAllowMultiline(t *testing.T) {
	t.Parallel()
	const source = "const element = <img className={`\n      d  c\n      b  a\n    `} />;"
	off := false

	rule_testing.ExpectFixedSource(t,
		rule_testing.Run(t, NoUnnecessaryWhitespace, "Component.tsx", source),
		"const element = <img className={`\n      d c\n      b a\n    `} />;")
	rule_testing.ExpectFixedSource(t,
		rule_testing.RunWithOptions(t, NoUnnecessaryWhitespace, "Component.tsx", source,
			NoUnnecessaryWhitespaceOptions{AllowMultiline: &off}),
		"const element = <img className={`d c b a`} />;")

	// A wrapped list with nothing else wrong is clean by default and a finding with the option off.
	const wrapped = "const element = <div className={`flex\n    items-center`} />;"
	rule_testing.ExpectClean(t, rule_testing.Run(t, NoUnnecessaryWhitespace, "Component.tsx", wrapped))
	rule_testing.ExpectFixedSource(t,
		rule_testing.RunWithOptions(t, NoUnnecessaryWhitespace, "Component.tsx", wrapped,
			NoUnnecessaryWhitespaceOptions{AllowMultiline: &off}),
		"const element = <div className={`flex items-center`} />;")
}
