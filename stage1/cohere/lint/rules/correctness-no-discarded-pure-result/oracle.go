//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/nexus"
)

func oracleWave17CorrectnessNoDiscardedPureResult() rule.Rule {
	return rules.CorrectnessNoDiscardedPureResult
}
func oracleWave17CorrectnessNoDiscardedPureResultOptions(fields []string) any {
	if len(fields) < 6 || fields[5] == "" || fields[5] == "null" {
		return nil
	}
	panic("nexus/correctness-no-discarded-pure-result has no options")
}
