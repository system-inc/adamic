//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/core"
)

func oracleRequireDescription() rule.Rule { return rules.RequireDescription }
func oracleRequireDescriptionOptions(fields []string) any {
	if len(fields) < 6 || fields[5] == "" || fields[5] == "null" {
		return nil
	}
	options, err := rules.DecodeRequireDescriptionOptions([]byte(fields[5]))
	if err != nil {
		panic(err)
	}
	return options
}
