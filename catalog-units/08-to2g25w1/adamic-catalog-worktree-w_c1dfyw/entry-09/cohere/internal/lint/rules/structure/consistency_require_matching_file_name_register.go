package structure

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers `react-component-require-matching-file-name` from a file of its own, so the rule is
// three owned files and no shared edit.
func init() {
	rule.Register(rule.Registration{Rule: ConsistencyRequireMatchingFileName})
}
