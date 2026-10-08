//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/core"
)

func oracleWave17NoThrowLiteral() rule.Rule { return rules.NoThrowLiteral }
func oracleWave17NoThrowLiteralOptions(fields []string) any {
	if len(fields) < 6 || fields[5] == "" || fields[5] == "null" {
		return nil
	}
	panic("no-throw-literal has no options")
}
