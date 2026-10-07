//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/typescript"
)

func oracleDefaultParamLast() rule.Rule                 { return rules.DefaultParamLast }
func oracleDefaultParamLastOptions(fields []string) any { return nil }
