//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/typescript"
)

func oracleCheckerQualifier() rule.Rule                 { return rules.NoUnnecessaryQualifier }
func oracleCheckerQualifierOptions(fields []string) any { return nil }
