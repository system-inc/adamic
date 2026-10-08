//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/core"
)

func oracleNoInlineComments() rule.Rule { return rules.NoInlineComments }
func oracleNoInlineCommentsOptions(fields []string) any {
	if len(fields) < 6 || fields[1] == "all" || fields[5] == "" || fields[5] == "null" {
		return nil
	}
	options, err := rules.DecodeNoInlineCommentsOptions([]byte(fields[5]))
	if err != nil {
		panic(err)
	}
	return options
}
