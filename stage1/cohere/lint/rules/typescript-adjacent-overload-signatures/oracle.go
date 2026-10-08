//go:build lintoracle

package main

import (
	"encoding/json"
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/typescript"
)

func oracleAdjacentOverloadSignatures() rule.Rule { return rules.AdjacentOverloadSignatures }

// This upstream rule has an empty options schema and ignores options in Run.
// Decode supplied JSON rather than silently dropping the manifest field.
func oracleAdjacentOverloadSignaturesOptions(fields []string) any {
	var options struct{}
	if len(fields) > 5 && fields[5] != "" {
		if err := json.Unmarshal([]byte(fields[5]), &options); err != nil {
			panic(err)
		}
	}
	return options
}
