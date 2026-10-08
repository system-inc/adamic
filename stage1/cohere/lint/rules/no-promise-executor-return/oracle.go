//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/core"
)

func oracleWave21PromiseReturn() rule.Rule { return rules.NoPromiseExecutorReturn }
func oracleWave21PromiseReturnOptions(fields []string) any {
	var raw []byte
	if len(fields) > 5 {
		raw = []byte(fields[5])
	}
	options, err := rules.DecodeNoPromiseExecutorReturnOptions(raw)
	if err != nil {
		panic(err)
	}
	return options
}
