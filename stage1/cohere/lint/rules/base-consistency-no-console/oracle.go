//go:build lintoracle

// Built only through the cohere overlay, so the upstream rule remains the oracle.
package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	base "github.com/system-inc/cohere/internal/lint/rules/base"
)

func oracleConsistencyNoConsole() rule.Rule                 { return base.ConsistencyNoConsole }
func oracleConsistencyNoConsoleOptions(fields []string) any { return nil }
