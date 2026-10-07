//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/core"
)

func oracleDefaultCase() rule.Rule { return rules.DefaultCase }
func oracleDefaultCaseOptions(fields []string) any {
	raw := ""
	if len(fields) > 5 {
		raw = fields[5]
	}
	options, err := rules.DecodeDefaultCaseOptions([]byte(raw))
	if err != nil {
		panic(err)
	}
	return options
}
