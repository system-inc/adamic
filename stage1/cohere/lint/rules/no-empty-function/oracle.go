//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/core"
)

func oracleNoEmptyFunction() rule.Rule { return rules.NoEmptyFunction }
func oracleNoEmptyFunctionOptions(fields []string) any {
	options, err := rules.DecodeNoEmptyFunctionOptions([]byte(fields[5]))
	if err != nil {
		panic(err)
	}
	return options
}
