package typescript

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers @typescript-eslint/strict-void-return.
//
// The Decode entry names the rule's own decoder. Its one option defaults to FALSE, so the generic
// helper would be correct by coincidence; the hand-rolled one keeps an absent key distinguishable
// from an explicit false.
func init() {
	rule.Register(rule.Registration{
		Rule:   StrictVoidReturn,
		Decode: DecodeStrictVoidReturnOptions,
	})
}
