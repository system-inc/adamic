//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/typescript"
)

func oracleNoUnnecessaryParameterPropertyAssignment() rule.Rule {
	return rules.NoUnnecessaryParameterPropertyAssignment
}
func oracleNoUnnecessaryParameterPropertyAssignmentOptions(fields []string) any { return nil }
