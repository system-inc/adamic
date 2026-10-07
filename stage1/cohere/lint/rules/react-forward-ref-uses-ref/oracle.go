//go:build lintoracle

// Built only through the cohere overlay, so the upstream rule remains the oracle.
package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/react"
)

func oracleForwardRefUsesRef() rule.Rule                 { return rules.ForwardRefUsesRef }
func oracleForwardRefUsesRefOptions(fields []string) any { return nil }
