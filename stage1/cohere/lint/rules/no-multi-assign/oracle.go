//go:build lintoracle

package main

import (
	"encoding/json"
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/core"
)

func oracleNoMultiAssign() rule.Rule { return rules.NoMultiAssign }

// All-rule rows also carry options belonging to other rules. Decode the fields
// this rule owns into its upstream type, as the unified adapters do.
func oracleNoMultiAssignOptions(fields []string) any {
	var options rules.NoMultiAssignOptions
	if len(fields) > 5 && fields[5] != "" && fields[5] != "null" {
		if err := json.Unmarshal([]byte(fields[5]), &options); err != nil {
			panic(err)
		}
	}
	return options
}
