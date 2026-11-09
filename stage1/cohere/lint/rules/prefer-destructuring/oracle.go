//go:build lintoracle

package main

import (
	"encoding/json"
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/core"
)

func oraclePreferDestructuring(options any) rule.Bound {
	return bindOracle(rules.PreferDestructuring, options)
}

// Captures carry the shipped decoder's typed result, not its raw schema tuple.
func oraclePreferDestructuringOptions(fields []string) any {
	enabled := true
	decoded := rules.PreferDestructuringOptions{
		VariableDeclarator:   rules.PreferDestructuringKindOptions{Array: &enabled, Object: &enabled},
		AssignmentExpression: rules.PreferDestructuringKindOptions{Array: &enabled, Object: &enabled},
	}
	if fields[5] != "" && fields[5] != "null" {
		decoded = rules.PreferDestructuringOptions{}
		if err := json.Unmarshal([]byte(fields[5]), &decoded); err != nil {
			panic(err)
		}
	}
	return decoded
}
