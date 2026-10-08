//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/nexus"
)

func oracleWave17CorrectnessNoDiscardedOutcome() rule.Rule {
	return rules.CorrectnessNoDiscardedOutcome
}
func oracleWave17CorrectnessNoDiscardedOutcomeOptions(fields []string) any {
	if len(fields) < 6 || fields[5] == "" || fields[5] == "null" {
		return nil
	}
	panic("nexus/correctness-no-discarded-outcome has no options")
}
