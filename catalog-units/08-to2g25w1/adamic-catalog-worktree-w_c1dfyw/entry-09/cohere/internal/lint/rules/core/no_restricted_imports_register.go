package core

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers no-restricted-imports. DecodeOptionList rather than Decode, because upstream's
// option surface is variadic: each restricted path is its own element, `["error", "fs", "os"]`, or a
// single element is the `{paths, patterns}` object.
func init() {
	rule.Register(rule.Registration{
		Rule:             NoRestrictedImports,
		DecodeOptionList: DecodeNoRestrictedImportsOptions,
	})
}
