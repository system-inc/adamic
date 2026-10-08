//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/react"
)

func oracleWave17UnsupportedSyntax() rule.Rule { return rules.UnsupportedSyntax }
func oracleWave17UnsupportedSyntaxOptions(fields []string) any {
	if len(fields) < 6 || fields[5] == "" || fields[5] == "null" {
		return nil
	}
	panic("react-hooks/unsupported-syntax has no options")
}
