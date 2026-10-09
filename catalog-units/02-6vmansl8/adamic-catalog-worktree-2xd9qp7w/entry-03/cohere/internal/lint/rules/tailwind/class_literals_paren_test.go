package tailwind

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// Parentheses are real nodes in this AST, and every rule in this package missed them.
//
// `className={('flex flex')}` is legal TSX meaning exactly what the unparenthesized form means, and
// all three shipped rules reported nothing on it: the readers unwrapped by the kinds they expected
// and walked past the parenthesis. Nothing failed, no fixture caught it, and the tree stayed green,
// which is the signature of an under-report rather than a bug.
//
// These are permanent fixtures rather than the throwaway probe that found it, because a defect that
// produced silence once will produce silence again.
func TestParenthesizedClassExpressionsAreRead(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name    string
		source  string
		subject func(*testing.T, string) rule_testing.Result
	}{
		{
			name:   "duplicate inside parentheses",
			source: `const element = <div className={('flex flex')} />;`,
			subject: func(t *testing.T, source string) rule_testing.Result {
				return rule_testing.Run(t, NoDuplicateClasses, "Component.tsx", source)
			},
		},
		{
			// The depth is the author's choice, so the unwrap has to recurse.
			name:   "duplicate inside nested parentheses",
			source: `const element = <div className={(('flex flex'))} />;`,
			subject: func(t *testing.T, source string) rule_testing.Result {
				return rule_testing.Run(t, NoDuplicateClasses, "Component.tsx", source)
			},
		},
		{
			name:   "duplicate inside a parenthesized call argument",
			source: `const merged = cn(('flex flex'));`,
			subject: func(t *testing.T, source string) rule_testing.Result {
				return rule_testing.Run(t, NoDuplicateClasses, "Component.tsx", source)
			},
		},
		{
			name:   "glued fragment inside parentheses",
			source: "const element = <div className={(`px-${size}`)} />;",
			subject: func(t *testing.T, source string) rule_testing.Result {
				return rule_testing.Run(t, NoConcatenatedClasses, "Component.tsx", source)
			},
		},
		{
			name:   "doubled space inside parentheses",
			source: `const element = <div className={('flex  items-center')} />;`,
			subject: func(t *testing.T, source string) rule_testing.Result {
				return rule_testing.Run(t, NoUnnecessaryWhitespace, "Component.tsx", source)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := testCase.subject(t, testCase.source)
			if len(result.Diagnostics) == 0 {
				t.Fatalf("parentheses hid the finding in %s", testCase.source)
			}
		})
	}
}

// The other direction: stripping parentheses must not make a rule fire on correct code.
//
// The finding that prompted this warned that `ast.SkipParentheses` is wrong for a rule whose verdict
// depends on the parse shape, and it is the house pattern in more than thirty rules here. These
// rules read a string's contents, which parentheses cannot change, so stripping is safe for them
// specifically. This pins that it stays a read and never becomes a rewrite.
func TestParenthesesDoNotCauseFalseFindings(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name    string
		source  string
		subject func(*testing.T, string) rule_testing.Result
	}{
		{
			name:   "clean classes inside parentheses",
			source: `const element = <div className={('flex items-center')} />;`,
			subject: func(t *testing.T, source string) rule_testing.Result {
				return rule_testing.Run(t, NoDuplicateClasses, "Component.tsx", source)
			},
		},
		{
			name:   "clean template inside parentheses",
			source: "const element = <div className={(`flex ${extra}`)} />;",
			subject: func(t *testing.T, source string) rule_testing.Result {
				return rule_testing.Run(t, NoConcatenatedClasses, "Component.tsx", source)
			},
		},
		{
			name:   "single spaces inside parentheses",
			source: `const element = <div className={('flex items-center')} />;`,
			subject: func(t *testing.T, source string) rule_testing.Result {
				return rule_testing.Run(t, NoUnnecessaryWhitespace, "Component.tsx", source)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, testCase.subject(t, testCase.source))
		})
	}
}
