//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/nexus"
)

func oracleCorrectnessNoDiscardedPureResult() rule.Rule {
	return rules.CorrectnessNoDiscardedPureResult
}
func oracleCorrectnessNoDiscardedPureResultOptions(fields []string) any { return nil }
