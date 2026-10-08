//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/next"
)

func oracleWave17NoAsyncClientComponent() rule.Rule { return rules.NoAsyncClientComponent }
func oracleWave17NoAsyncClientComponentOptions(fields []string) any {
	if len(fields) < 6 || fields[5] == "" || fields[5] == "null" {
		return nil
	}
	panic("@next/next/no-async-client-component has no options")
}
