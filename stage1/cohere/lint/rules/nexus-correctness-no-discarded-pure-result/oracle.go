//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/nexus"
)

func oracleWave21PureResult() rule.Rule                 { return rules.CorrectnessNoDiscardedPureResult }
func oracleWave21PureResultOptions(fields []string) any { return nil }
