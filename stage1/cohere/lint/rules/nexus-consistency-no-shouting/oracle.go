//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/nexus"
)

func oracleNexusConsistencyNoShouting(options any) rule.Bound {
	return bindOracle(rules.ConsistencyNoShouting, options)
}

// Cohere no longer accepts an allow list. Keep the case bag out of the rule.
func oracleNexusConsistencyNoShoutingOptions(fields []string) any { return rule.NoOptions{} }
