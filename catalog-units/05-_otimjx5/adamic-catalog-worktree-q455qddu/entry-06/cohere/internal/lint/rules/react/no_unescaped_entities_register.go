package react

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers this package's `no-unescaped-entities` rule.
//
// A file of its own rather than a line in `register.go`, so a port is three owned files and no
// shared edit. The decoder is hand-written rather than `rule.DecodeOptionsInto` because the
// `forbid` array is heterogeneous and has to be distinguishable as absent; see
// DecodeNoUnescapedEntitiesOptions.
func init() {
	rule.Register(rule.Registration{Rule: NoUnescapedEntities, Decode: DecodeNoUnescapedEntitiesOptions})
}
