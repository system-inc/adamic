//go:build lintoracle

// Built only through the cohere overlay, so the upstream rule remains the oracle.
package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/core"
)

func oracleNoEmpty() rule.Rule { return rules.NoEmpty }
func oracleNoEmptyOptions(fields []string) any {
	return rules.NoEmptyOptions{AllowEmptyCatch: fields[4] == "true"}
}
