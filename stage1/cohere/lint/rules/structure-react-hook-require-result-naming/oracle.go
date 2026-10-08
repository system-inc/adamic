//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/structure"
)

func oracleStructureReactHookRequireResultNaming() rule.Rule {
	return rules.ReactHookRequireResultNaming
}
func oracleStructureReactHookRequireResultNamingOptions(fields []string) any { return nil }
