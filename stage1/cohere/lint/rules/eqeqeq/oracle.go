//go:build lintoracle

// Built only through the cohere overlay, so the upstream rule remains the oracle.
package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/core"
)

func oracleEqeqeq() rule.Rule { return rules.Eqeqeq }
func oracleEqeqeqOptions(fields []string) any {
	return rules.EqeqeqOptions{Mode: rules.EqeqeqMode(fields[2]), Null: rules.EqeqeqNullPolicy(fields[3])}
}
