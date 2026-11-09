//go:build lintoracle

// Built only through the cohere overlay, so the upstream rule remains the oracle.
package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	nexus "github.com/system-inc/cohere/internal/lint/rules/nexus"
)

func oracleConsistencyRequireTypeSuffix(options any) rule.Bound {
	return bindOracle(nexus.ConsistencyRequireTypeSuffix, options)
}
func oracleConsistencyRequireTypeSuffixOptions(fields []string) any { return nil }
