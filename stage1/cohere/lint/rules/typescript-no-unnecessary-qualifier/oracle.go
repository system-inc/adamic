//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/typescript"
)

func oracleWave22NoUnnecessaryQualifier() rule.Rule                 { return rules.NoUnnecessaryQualifier }
func oracleWave22NoUnnecessaryQualifierOptions(fields []string) any { return nil }
