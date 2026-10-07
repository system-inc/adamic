//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/base"
)

func oracleContextAccess() rule.Rule { return rules.SecurityRequireContextAccess }
func oracleContextAccessOptions(fields []string) any {
	if len(fields) < 6 || fields[5] == "" {
		return nil
	}
	options, err := rules.DecodeContextRequiresAccessOptions([]byte(fields[5]))
	if err != nil {
		panic(err)
	}
	return options
}
