package typescript

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers @typescript-eslint/init-declarations.
//
// Its own file rather than a line in the package's register.go: that shared list is one file every
// concurrent port has to edit, so two ports landing at once conflict on a file neither is really
// changing.
//
// DecodeOptionList rather than Decode, because upstream's option surface is a positional list whose
// first element is a bare mode string and whose second carries `ignoreForLoopInit`:
// `["error", "never", {"ignoreForLoopInit": true}]`. The generic struct decoder cannot express that.
func init() {
	rule.Register(rule.Registration{
		Rule:             InitDeclarations,
		DecodeOptionList: DecodeInitDeclarationsOptions,
	})
}
