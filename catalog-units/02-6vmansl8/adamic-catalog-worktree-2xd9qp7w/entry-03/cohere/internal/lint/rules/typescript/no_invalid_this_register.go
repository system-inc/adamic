package typescript

import "github.com/system-inc/cohere/internal/lint/rule"

func init() {
	rule.Register(rule.Registration{
		Rule:   NoInvalidThis,
		Decode: DecodeNoInvalidThisOptions,
	})
}
