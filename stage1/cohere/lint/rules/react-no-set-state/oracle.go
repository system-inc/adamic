//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/react"
)

func oracleReactNoSetState() rule.Rule                 { return rules.NoSetState }
func oracleReactNoSetStateOptions(fields []string) any { return nil }
