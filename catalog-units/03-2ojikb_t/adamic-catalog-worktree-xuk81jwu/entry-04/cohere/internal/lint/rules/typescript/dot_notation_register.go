package typescript

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers @typescript-eslint/dot-notation.
//
// Its own file rather than a line in the package's register.go, so two ports landing at once do not
// conflict on a file neither is really changing. The decoder is hand-written because `allowKeywords`
// defaults to TRUE; see DecodeDotNotationOptions.
func init() {
	rule.Register(rule.Registration{Rule: DotNotation, Decode: DecodeDotNotationOptions})
}
