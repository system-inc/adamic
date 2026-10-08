//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/typescript"
)

func oraclePreferReduceTypeParameter() rule.Rule                 { return rules.PreferReduceTypeParameter }
func oraclePreferReduceTypeParameterOptions(fields []string) any { return nil }
