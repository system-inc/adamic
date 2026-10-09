//go:build lintoracle

package main

import (
	"encoding/json"
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/core"
)

func oracleNoExAssign() rule.Rule { return rules.NoExAssign }

// Upstream has no options type and reads no options. Validate supplied JSON
// and return an explicit empty value so the harness never silently drops it.
func oracleNoExAssignOptions(fields []string) any {
	var options struct{}
	if fields[5] != "" {
		if err := json.Unmarshal([]byte(fields[5]), &options); err != nil {
			panic(err)
		}
	}
	return options
}
