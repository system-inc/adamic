//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/next"
)

func oracleWave17NoDuplicateHead() rule.Rule { return rules.NoDuplicateHead }
func oracleWave17NoDuplicateHeadOptions(fields []string) any {
	if len(fields) < 6 || fields[5] == "" || fields[5] == "null" {
		return nil
	}
	panic("@next/next/no-duplicate-head has no options")
}
