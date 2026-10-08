//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/nexus"
)

func oracleCorrectnessNoUnclearedRaceTimeout() rule.Rule {
	return rules.CorrectnessNoUnclearedRaceTimeout
}
func oracleCorrectnessNoUnclearedRaceTimeoutOptions(fields []string) any { return nil }
