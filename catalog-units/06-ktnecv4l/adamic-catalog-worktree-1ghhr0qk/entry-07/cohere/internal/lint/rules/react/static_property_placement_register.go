package react

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers this package's `static-property-placement` rule.
//
// DecodeOptionList rather than Decode, and hand-written rather than `rule.DecodeOptionsInto`,
// because upstream's option surface is a POSITIONAL pair whose two slots have different types:
// `["error", "property assignment", {"displayName": "static getter"}]`. See
// DecodeStaticPropertyPlacementOptions.
func init() {
	rule.Register(rule.Registration{
		Rule:             StaticPropertyPlacement,
		DecodeOptionList: DecodeStaticPropertyPlacementOptions,
	})
}
