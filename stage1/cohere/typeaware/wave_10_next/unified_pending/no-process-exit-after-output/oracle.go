//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/nexus"
)

func oracleCorrectnessNoProcessExitAfterOutput() rule.Rule {
	return rules.CorrectnessNoProcessExitAfterOutput
}
func oracleCorrectnessNoProcessExitAfterOutputOptions(fields []string) any { return nil }
