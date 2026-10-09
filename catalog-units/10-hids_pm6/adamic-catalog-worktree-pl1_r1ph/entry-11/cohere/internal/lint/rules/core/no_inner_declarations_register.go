package core

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers no-inner-declarations. DecodeOptionList rather than Decode, because upstream's
// option surface is a two-element list: `["error", "both", {"blockScopedFunctions": "disallow"}]`.
func init() {
	rule.Register(rule.Registration{
		Rule:             NoInnerDeclarations,
		DecodeOptionList: DecodeNoInnerDeclarationsOptions,
	})
}
