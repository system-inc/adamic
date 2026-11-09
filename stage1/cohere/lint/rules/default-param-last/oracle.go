//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/core"
)

func oracleDefaultParamLastCore(options any) rule.Bound {
	return bindOracle(rules.DefaultParamLast, options)
}
func oracleDefaultParamLastCoreOptions(fields []string) any { return nil }
