//go:build lintoracle

package main

import (
	"encoding/json"
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/core"
)

func oracleNoUnsafeFinally() rule.Rule { return rules.NoUnsafeFinally }

// The upstream rule has no options type and ignores its options argument.
// Decode supplied JSON nevertheless so an options row cannot reach a nil adapter.
func oracleNoUnsafeFinallyOptions(fields []string) any {
	if len(fields) > 5 && fields[5] != "" && fields[5] != "null" {
		var options map[string]json.RawMessage
		if err := json.Unmarshal([]byte(fields[5]), &options); err != nil {
			panic(err)
		}
		return options
	}
	return nil
}
