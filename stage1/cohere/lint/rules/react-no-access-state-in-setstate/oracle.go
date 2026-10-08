//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/react"
)

func oracleReactNoAccessStateInSetstate() rule.Rule                 { return rules.NoAccessStateInSetstate }
func oracleReactNoAccessStateInSetstateOptions(fields []string) any { return nil }
