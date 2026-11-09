//go:build lintoracle

// Built only through the cohere overlay, so the upstream rule remains the oracle.
package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	base "github.com/system-inc/cohere/internal/lint/rules/base"
)

func oracleConsistencyNoConsole(options any) rule.Bound {
	return bindOracle(base.ConsistencyNoConsole, options)
}
func oracleConsistencyNoConsoleOptions(fields []string) any { return nil }
