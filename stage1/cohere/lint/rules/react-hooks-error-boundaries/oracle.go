//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/react"
)

func oracleReactHooksErrorBoundaries() rule.Rule                 { return rules.ErrorBoundaries }
func oracleReactHooksErrorBoundariesOptions(fields []string) any { return rules.CompilerRuleOptions{} }
