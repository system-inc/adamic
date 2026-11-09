//go:build lintoracle

// Built only through the cohere overlay, so the upstream rule remains the oracle.
package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/react"
)

func oracleNoIsMounted(options any) rule.Bound     { return bindOracle(rules.NoIsMounted, options) }
func oracleNoIsMountedOptions(fields []string) any { return nil }
