//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/react"
)

func oracleReactHooksVoidUseMemo() rule.Rule                 { return rules.VoidUseMemo }
func oracleReactHooksVoidUseMemoOptions(fields []string) any { return rules.CompilerRuleOptions{} }
