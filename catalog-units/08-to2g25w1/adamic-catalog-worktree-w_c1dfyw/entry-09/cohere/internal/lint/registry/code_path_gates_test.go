package registry

import "github.com/system-inc/cohere/internal/lint/rules/core"

// The ESLint corpus replays here, so the gated rules check their gates here too: a root a gate skips
// is built anyway and must hold nothing the rule could report.
func init() {
	core.CheckCodePathGates = true
}
