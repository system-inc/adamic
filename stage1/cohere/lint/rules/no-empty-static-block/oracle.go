//go:build lintoracle

// Built only through the cohere overlay, so the upstream rule remains the oracle.
package main

import (
	"encoding/json"
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/core"
)

func oracleNoEmptyStaticBlock() rule.Rule { return rules.NoEmptyStaticBlock }

// Upstream has no configurable options or options type. Decode its empty object
// explicitly so an options-bearing row cannot reach a nil adapter result.
func oracleNoEmptyStaticBlockOptions(fields []string) any {
	var decoded struct{}
	if len(fields) > 5 && fields[5] != "" {
		if err := json.Unmarshal([]byte(fields[5]), &decoded); err != nil {
			panic(err)
		}
	}
	return decoded
}
