//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/structure"
)

func oracleStructureReactComponentRequireNamedExport() rule.Rule {
	return rules.ReactComponentRequireNamedExport
}
func oracleStructureReactComponentRequireNamedExportOptions(fields []string) any { return nil }
