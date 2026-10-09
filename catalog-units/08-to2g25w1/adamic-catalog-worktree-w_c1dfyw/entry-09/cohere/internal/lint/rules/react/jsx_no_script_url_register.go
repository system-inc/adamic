package react

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers this package's `jsx-no-script-url` rule.
//
// A file of its own rather than a line in `register.go`, so a port is three owned files and no
// shared edit.
//
// DecodeOptionList rather than Decode, and hand-rolled rather than `rule.DecodeOptionsInto`, because
// upstream's option surface is POSITIONAL across two slots whose shapes differ:
// `["error", [{"name": "Foo", "props": ["to"]}], {"includeFromSettings": false}]`. The decoder also
// has to answer an empty list with the built-in `a` to `href` pair rather than an error, since an
// unconfigured rule still checks anchors.
func init() {
	rule.Register(rule.Registration{
		Rule:             JsxNoScriptUrl,
		DecodeOptionList: DecodeJsxNoScriptUrlOptions,
	})
}
