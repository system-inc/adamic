//go:build lintoracle

// Built only through the cohere overlay, so the upstream rule remains the oracle.
package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	nexus "github.com/system-inc/cohere/internal/lint/rules/nexus"
)

func oracleConsistencyNoEnum() rule.Rule                 { return nexus.ConsistencyNoEnum }
func oracleConsistencyNoEnumOptions(fields []string) any { return nil }
