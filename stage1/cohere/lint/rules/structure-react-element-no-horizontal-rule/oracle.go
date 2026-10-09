//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/structure"
)

func oracleStructureReactElementNoHorizontalRule() rule.Rule {
	return rules.ReactElementNoHorizontalRule
}
func oracleStructureReactElementNoHorizontalRuleOptions(fields []string) any { return nil }
