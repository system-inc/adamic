package nexus

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers nexus/correctness-no-load-time-import-meta-path.
//
// No decoder: the rule takes no options. Whether a read runs at load is a fact of the syntax, and the
// one exempt shape, the main-module guard, is exempt because it is harmless everywhere.
func init() {
	rule.Register(rule.Registration{Rule: CorrectnessNoLoadTimeImportMetaPath})
}
