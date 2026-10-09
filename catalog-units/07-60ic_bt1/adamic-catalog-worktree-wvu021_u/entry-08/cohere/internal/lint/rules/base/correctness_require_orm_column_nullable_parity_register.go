package base

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers base/correctness-require-orm-column-nullable-parity.
//
// No Decode entry: `schema: []` in the source and the rule reads no options.
func init() {
	rule.Register(rule.Registration{Rule: CorrectnessRequireOrmColumnNullableParity})
}
