//go:build lintoracle

package main

import (
	"encoding/json"
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/tailwind"
)

func oracleEnforceConsistentImportantPosition() rule.Rule {
	return rules.EnforceConsistentImportantPosition
}
func oracleEnforceConsistentImportantPositionOptions(fields []string) any {
	if len(fields) > 5 && fields[5] != "" && fields[5] != "null" {
		var decoded rules.EnforceConsistentImportantPositionOptions
		if err := json.Unmarshal([]byte(fields[5]), &decoded); err != nil {
			panic(err)
		}
		return decoded
	}
	return nil
}
