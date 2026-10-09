package react

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers this package's `sort-default-props` rule.
func init() {
	rule.Register(rule.Registration{Rule: SortDefaultProps, Decode: DecodeSortDefaultPropsOptions})
}
