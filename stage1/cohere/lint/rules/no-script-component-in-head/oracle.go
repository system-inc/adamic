//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/next"
)

func oracleWave17NoScriptComponentInHead() rule.Rule { return rules.NoScriptComponentInHead }
func oracleWave17NoScriptComponentInHeadOptions(fields []string) any {
	if len(fields) < 6 || fields[5] == "" || fields[5] == "null" {
		return nil
	}
	panic("@next/next/no-script-component-in-head has no options")
}
