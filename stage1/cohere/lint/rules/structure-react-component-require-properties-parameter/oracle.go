//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/structure"
)

func oracleStructureReactComponentRequirePropertiesParameter() rule.Rule {
	return rules.ReactComponentRequirePropertiesParameter
}
func oracleStructureReactComponentRequirePropertiesParameterOptions(fields []string) any { return nil }
