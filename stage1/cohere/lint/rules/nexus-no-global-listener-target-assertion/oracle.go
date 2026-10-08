//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/nexus"
)

func oracleWave14CorrectnessNoGlobalListenerTargetAssertion() rule.Rule {
	return rules.CorrectnessNoGlobalListenerTargetAssertion
}
func oracleWave14CorrectnessNoGlobalListenerTargetAssertionOptions(fields []string) any { return nil }
