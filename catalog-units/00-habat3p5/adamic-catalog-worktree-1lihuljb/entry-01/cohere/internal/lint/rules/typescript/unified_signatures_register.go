package typescript

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers @typescript-eslint/unified-signatures.
//
// Its own file rather than a line in the package's register.go, so two ports landing at once do not
// conflict on a file neither is really changing.
//
// `DecodeOptionsInto` rather than a hand-written decoder: both of upstream's options default to
// `false`, which is Go's zero value for a bool, so a rule handed an empty settings struct behaves
// exactly as a rule handed upstream's defaults. That is the condition a generic decode needs, and it
// is worth stating because it is NOT true of every optioned rule in this package.
func init() {
	rule.Register(rule.Registration{
		Rule:   UnifiedSignatures,
		Decode: rule.DecodeOptionsInto[UnifiedSignaturesOptions](),
	})
}
