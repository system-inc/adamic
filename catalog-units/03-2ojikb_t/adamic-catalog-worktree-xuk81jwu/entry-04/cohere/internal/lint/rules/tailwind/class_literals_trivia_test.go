package tailwind

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// A class literal's range starts at its opening quote, not inside the trivia before it.
//
// classLiteralFrom took the node's Loc, whose start includes leading whitespace, so for a literal on
// its own line or after a comma the "strip the quotes" step cut one character too early: the fix
// replaced the opening quote, the file stopped parsing, and the edit engine refused that file's
// whole batch, every other rule's fixes included. Measured by @system_cohere_format_markdown on
// perturbed input: 174 files in ahra and 175 in www-phi-health (1,755 and 1,932 refused fixes).
func TestClassLiteralFixesKeepTheirQuotesAfterTrivia(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name, source, want string
	}{
		{
			name:   "literal on its own line",
			source: "const merged = cn(\n    'flex flex',\n);\n",
			want:   "const merged = cn(\n    'flex',\n);\n",
		},
		{
			name:   "literal after a comma",
			source: "const merged = cn('p-2', 'flex flex');\n",
			want:   "const merged = cn('p-2', 'flex');\n",
		},
		{
			// The control: no trivia before the literal, which was always right.
			name:   "literal with no trivia before it",
			source: "const merged = cn('flex flex');\n",
			want:   "const merged = cn('flex');\n",
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
