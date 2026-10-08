//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/base"
)

func oracleContextAccess() rule.Rule { return rules.SecurityRequireContextAccess }
func oracleContextAccessOptions(fields []string) any {
	// Shared all-rule rows carry legacy options for other rules; security options belong to selected rows.
	if len(fields) < 6 || fields[1] != "base/security-require-context-access" || fields[5] == "" {
		return nil
	}
	options, err := rules.DecodeContextRequiresAccessOptions([]byte(fields[5]))
	if err != nil {
		panic(err)
	}
	return options
}
