//go:build lintoracle

package main

import (
	"encoding/json"
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/core"
)

func oracleOperatorAssignment() rule.Rule { return rules.OperatorAssignment }
func oracleOperatorAssignmentOptions(fields []string) any {
	options := rules.DefaultOperatorAssignmentSettings()
	if len(fields) > 5 && fields[5] != "" {
		raw := []byte(fields[5])
		if raw[0] == '"' {
			decoded, err := rules.DecodeOperatorAssignmentOptions(raw)
			if err != nil {
				panic(err)
			}
			return decoded
		}
		if err := json.Unmarshal(raw, &options); err != nil {
			panic(err)
		}
	}
	return options
}
