package typescript

import "github.com/system-inc/cohere/internal/lint/rule"

func init() {
	rule.Register(rule.Registration{
		Rule:   ConsistentTypeDefinitions,
		Decode: DecodeConsistentTypeDefinitionsOptions,
	})
}
