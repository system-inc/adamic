//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/nexus"
)

func oracleConsistencyNoHandRolledDelay() rule.Rule                 { return rules.ConsistencyNoHandRolledDelay }
func oracleConsistencyNoHandRolledDelayOptions(fields []string) any { return nil }
