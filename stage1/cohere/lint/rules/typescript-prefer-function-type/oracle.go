//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/typescript"
)

func oraclePreferFunctionType() rule.Rule                 { return rules.PreferFunctionType }
func oraclePreferFunctionTypeOptions(fields []string) any { return nil }
