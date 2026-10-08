//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/core"
)

func oracleCheckerPromiseReturn() rule.Rule { return rules.NoPromiseExecutorReturn }
func oracleCheckerPromiseReturnOptions(fields []string) any {
	if len(fields) <= 5 || fields[5] == "" || fields[5] == "null" {
		return nil
	}
	options, err := rules.DecodeNoPromiseExecutorReturnOptions([]byte(fields[5]))
	if err != nil {
		panic(err)
	}
	return options
}
