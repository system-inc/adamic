package base

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers base/correctness-require-orm-column-declare.
//
// No Decode entry, because the source rule's `schema` is `[]` and it reads no options.
func init() {
	rule.Register(rule.Registration{Rule: CorrectnessRequireOrmColumnDeclare})
}
