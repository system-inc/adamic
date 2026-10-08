//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/react"
)

func oracleWave17UseMemo() rule.Rule { return rules.UseMemo }
func oracleWave17UseMemoOptions(fields []string) any {
	if len(fields) < 6 || fields[5] == "" || fields[5] == "null" {
		return nil
	}
	panic("react-hooks/use-memo has no options")
}
