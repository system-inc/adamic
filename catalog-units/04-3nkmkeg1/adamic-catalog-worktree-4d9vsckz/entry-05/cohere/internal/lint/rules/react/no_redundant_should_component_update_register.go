package react

import "github.com/system-inc/cohere/internal/lint/rule"

// No decoder. Upstream's `meta.schema` is `[]`, confirmed against the installed build.
func init() {
	rule.Register(rule.Registration{Rule: NoRedundantShouldComponentUpdate})
}
