package boundaries

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers boundaries/dependencies.
//
// DecodeAt because its element patterns are paths relative to where the config was written, and
// RequiresOptions because without elements no file belongs to a layer and the rule decides nothing.
func init() {
	rule.Register(rule.Registration{
		Rule:            Dependencies,
		DecodeAt:        decodeDependenciesOptions,
		RequiresOptions: true,
	})
}
