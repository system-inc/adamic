package tailwind

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// Every string that can become the class value is read, wherever the expression puts it.
//
// Until #vf1hd6j the reader returned only the surface's own string. Prettier's Tailwind plugin
// sorted the strings in conditionals, `&&` and `??` operands, arrays and template holes, and cohere
// read none of them: 364 such literals in ahra and 299 in www-phi-health that no rule had ever seen.
// A template with no holes was excluded on purpose, over escape offsets that the fixers now guard
// themselves.
//
// Each case plants a repeat in exactly one position, so a finding there is proof that position is
// read, and the fix is asserted so the range is proven to land on the right string.
func TestNestedClassLiteralsAreRead(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name, source, want string
	}{
		{
			name:   "both branches of a conditional",
			source: `const element = <div className={open ? 'flex flex' : 'hidden hidden'} />;`,
			want:   `const element = <div className={open ? 'flex' : 'hidden'} />;`,
		},
		{
			name:   "the right side of &&",
			source: `const element = <div className={open && 'flex flex'} />;`,
			want:   `const element = <div className={open && 'flex'} />;`,
		},
		{
			name:   "both sides of ??",
			source: `const element = <div className={override ?? 'p-2 p-2'} />;`,
			want:   `const element = <div className={override ?? 'p-2'} />;`,
		},
		{
			name:   "array elements",
			source: `const merged = cn(['flex flex', open && 'p-2 p-2']);`,
			want:   `const merged = cn(['flex', open && 'p-2']);`,
		},
		{
			name:   "a string inside a template's hole",
			source: "const element = <div className={`flex ${open ? 'p-2 p-2' : ''}`} />;",
			want:   "const element = <div className={`flex ${open ? 'p-2' : ''}`} />;",
		},
		{
			name:   "a template with no holes",
			source: "const element = <div className={`flex flex`} />;",
			want:   "const element = <div className={`flex`} />;",
		},
		{
			name:   "a callee argument's conditional",
			source: `const merged = cn('p-2', open ? 'flex flex' : undefined);`,
			want:   `const merged = cn('p-2', open ? 'flex' : undefined);`,
		},
		{
			name:   "a variable's initializer behind as const",
			source: `const className = (open ? 'flex flex' : 'hidden') as const;`,
			want:   `const className = (open ? 'flex' : 'hidden') as const;`,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoDuplicateClasses, "Component.tsx", testCase.source)
			rule_testing.ExpectFixedSource(t, result, testCase.want)
		})
	}
}

// The other half: strings that are not a class value are not read, though the plugin reads them.
//
// The plugin's `sortInside` visits every string under the braces, so it sorts `'sm'` in a comparison
// too, which costs it nothing because a single word sorts to itself. Reading those strings here
// would send `'primary large'` from a comparison to `no-unknown-classes`, and a condition is not a
// class list. Each source below holds a repeat that only a wrong reading would find.
func TestNonValueStringsAreNotRead(t *testing.T) {
	t.Parallel()
	for _, testCase := range []struct {
		name, source string
	}{
		{
			name:   "a conditional's condition",
			source: `const element = <div className={size === 'sm sm' ? 'p-2' : 'p-4'} />;`,
		},
		{
			name:   "the left side of &&",
			source: `const element = <div className={'flex flex' && 'p-2'} />;`,
		},
		{
			name:   "an argument to a function nobody named a class callee",
			source: `const element = <div className={translate('hello hello')} />;`,
		},
		{
			name:   "a string concatenated into another",
			source: `const element = <div className={'flex flex' + suffix} />;`,
		},
		{
			name:   "a conditional on an unrelated attribute",
			source: `const element = <div title={open ? 'flex flex' : ''} />;`,
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoDuplicateClasses, "Component.tsx", testCase.source)
			rule_testing.ExpectClean(t, result)
		})
	}
}

// A string inside a template's hole keeps the whitespace at its edges.
//
// In `flex${open ? ' hidden' : ”}` the leading space is all that keeps `hidden` apart from `flex`
// once the value is substituted, so trimming it would fuse the two into a class that does not exist.
// The plugin's `canCollapseWhitespaceIn` decides from the same position and keeps it. The same
// string outside a template is trimmed, which is the discrimination that makes the first case mean
// something.
func TestWhitespaceAtAHoleLiteralsEdgeIsASeparator(t *testing.T) {
	t.Parallel()
	t.Run("inside a hole, the edge keeps one character", func(t *testing.T) {
		t.Parallel()
		result := rule_testing.Run(t, NoUnnecessaryWhitespace, "Component.tsx",
			"const element = <div className={`flex${open ? '   hidden' : ''}`} />;")
		rule_testing.ExpectFixedSource(t, result,
			"const element = <div className={`flex${open ? ' hidden' : ''}`} />;")
	})

	t.Run("inside a hole, a single edge space is no finding", func(t *testing.T) {
		t.Parallel()
		result := rule_testing.Run(t, NoUnnecessaryWhitespace, "Component.tsx",
			"const element = <div className={`flex${open ? ' hidden' : ''}`} />;")
		rule_testing.ExpectClean(t, result)
	})

	t.Run("outside a template, the same edge is padding", func(t *testing.T) {
		t.Parallel()
		result := rule_testing.Run(t, NoUnnecessaryWhitespace, "Component.tsx",
			"const element = <div className={open ? ' hidden' : ''} />;")
		rule_testing.ExpectFixedSource(t, result,
			"const element = <div className={open ? 'hidden' : ''} />;")
	})

	// The hole's text still reads as segments when a hole holds a string. They used to be
	// alternatives, and this doubled space was the casualty.
	t.Run("the template's own text is still read", func(t *testing.T) {
		t.Parallel()
		result := rule_testing.Run(t, NoUnnecessaryWhitespace, "Component.tsx",
			"const element = <div className={`flex  ${open ? 'hidden' : ''}`} />;")
		rule_testing.ExpectFixedSource(t, result,
			"const element = <div className={`flex ${open ? 'hidden' : ''}`} />;")
	})
}

// A template whose source is not its value is reported and not fixed.
//
// `\“ decodes to a backtick, so writing the decoded classes back over the source would close the
// template early, the file would stop parsing, and the engine would refuse every fix in it.
func TestAnEscapedTemplateIsReportedWithoutAFix(t *testing.T) {
	t.Parallel()
	result := rule_testing.Run(t, NoDuplicateClasses, "Component.tsx",
		"const element = <div className={`flex flex content-['\\`']`} />;")
	rule_testing.ExpectFindings(t, result, "duplicateClass")
	if len(result.Diagnostics[0].Fixes) != 0 {
		t.Fatalf("expected no fix over an escape, got %+v", result.Diagnostics[0].Fixes)
	}
}

// A template's runs are found at their tokens, so trivia in front of a delimiter does not shift
// them. Before, a template after `? ` or a span after `${ x }` read as untokenizable, and every rule
// that edits runs skipped it; the direct template, with nothing in front of its backtick, was fine.
func TestTemplateRunsAfterTriviaAreFixed(t *testing.T) {
	t.Parallel()
	for _, testCase := range []struct {
		name, source, want string
	}{
		{
			name:   "a template after a conditional's question mark",
			source: "const element = <div className={open ? `flex  block ${x}` : ''} />;",
			want:   "const element = <div className={open ? `flex block ${x}` : ''} />;",
		},
		{
			name:   "a run after a hole with spaces inside its braces",
			source: "const element = <div className={`flex ${ x }  block  gap-2`} />;",
			want:   "const element = <div className={`flex ${ x } block gap-2`} />;",
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoUnnecessaryWhitespace, "Component.tsx", testCase.source)
			rule_testing.ExpectFixedSource(t, result, testCase.want)
		})
	}
}

// A string in a template's hole loses its edge whitespace exactly where the template text beside the
// hole already supplies some, and keeps one character where it does not.
//
// The trimmed cases are quoted from Prettier's Tailwind plugin on 2026-10-02. The kept cases are
// where the plugin is wrong: it trims them too, so `flex${c ? '  block  ' : ”}` became
// `flex${c ? 'block' : ”}` and two classes fused into `flexblock`. See holeEdges.
func TestHoleStringEdgesTrimOnlyWhereTheTemplateSeparates(t *testing.T) {
	t.Parallel()
	for _, testCase := range []struct {
		name, source, want string
	}{
		{
			name:   "the gate's shape: text before ends with a space, nothing after",
			source: "const element = <i className={`size-2.5 transition-transform ${open ? ' rotate-90 ' : ''}`} />;",
			want:   "const element = <i className={`size-2.5 transition-transform ${open ? 'rotate-90' : ''}`} />;",
		},
		{
			name:   "spaces on both sides of the hole",
			source: "const element = <i className={`flex ${c ? '  block  ' : ''} grid`} />;",
			want:   "const element = <i className={`flex ${c ? 'block' : ''} grid`} />;",
		},
		{
			name:   "a hole at the template's start",
			source: "const element = <i className={`${c ? '  block  ' : ''} flex`} />;",
			want:   "const element = <i className={`${c ? 'block' : ''} flex`} />;",
		},
		{
			name:   "a template nested in a hole",
			source: "const element = <i className={`flex ${c ? `  block  ${d}` : ''}`} />;",
			want:   "const element = <i className={`flex ${c ? `block ${d}` : ''}`} />;",
		},
		{
			// The plugin writes `flex${c ? 'block' : ''}grid` here, which is `flexblockgrid`.
			name:   "glued on both sides, one character survives each edge",
			source: "const element = <i className={`flex${c ? '  block  ' : ''}grid`} />;",
			want:   "const element = <i className={`flex${c ? ' block ' : ''}grid`} />;",
		},
		{
			name:   "glued before, spaced after",
			source: "const element = <i className={`flex${c ? '  block  ' : ''} grid`} />;",
			want:   "const element = <i className={`flex${c ? ' block' : ''} grid`} />;",
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoUnnecessaryWhitespace, "Component.tsx", testCase.source)
			rule_testing.ExpectFixedSource(t, result, testCase.want)
		})
	}

	// And the already-right shapes stay silent, glued single spaces included.
	for _, source := range []string{
		"const element = <i className={`flex ${c ? 'block' : ''}`} />;",
		"const element = <i className={`flex${c ? ' block' : ''}`} />;",
		"const element = <i className={`${c ? 'block ' : ''}flex`} />;",
	} {
		rule_testing.ExpectClean(t, rule_testing.Run(t, NoUnnecessaryWhitespace, "Component.tsx", source))
	}
}
