//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/core"
)

func oracleDefaultParamLastCore() rule.Rule                 { return rules.DefaultParamLast }
func oracleDefaultParamLastCoreOptions(fields []string) any { return nil }
