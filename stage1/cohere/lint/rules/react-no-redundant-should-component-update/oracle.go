//go:build lintoracle

// Built only through the cohere overlay, so the upstream rule remains the oracle.
package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/react"
)

func oracleNoRedundantShouldComponentUpdate() rule.Rule {
	return rules.NoRedundantShouldComponentUpdate
}
func oracleNoRedundantShouldComponentUpdateOptions(fields []string) any { return nil }
