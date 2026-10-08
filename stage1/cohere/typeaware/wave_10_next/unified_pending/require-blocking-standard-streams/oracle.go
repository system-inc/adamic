//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/nexus"
)

func oracleCorrectnessRequireBlockingStandardStreams() rule.Rule {
	return rules.CorrectnessRequireBlockingStandardStreams
}
func oracleCorrectnessRequireBlockingStandardStreamsOptions(fields []string) any { return nil }
