//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/react"
)

func oracleReactNoUnusedState() rule.Rule                 { return rules.NoUnusedState }
func oracleReactNoUnusedStateOptions(fields []string) any { return nil }
