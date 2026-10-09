//go:build lintoracle

// Built only through the cohere overlay, so the upstream rule remains the oracle.
package main

import (
	"encoding/json"
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/core"
)

func oracleEqeqeq() rule.Rule { return rules.Eqeqeq }
func oracleEqeqeqOptions(fields []string) any {
	if len(fields) > 5 && fields[5] != "" && fields[5] != "null" {
		options := rules.EqeqeqOptions{Mode: rules.EqeqeqMode(fields[2]), Null: rules.EqeqeqNullPolicy(fields[3])}
		if err := json.Unmarshal([]byte(fields[5]), &options); err != nil {
			panic(err)
		}
		return options
	}

	return rules.EqeqeqOptions{Mode: rules.EqeqeqMode(fields[2]), Null: rules.EqeqeqNullPolicy(fields[3])}
}
