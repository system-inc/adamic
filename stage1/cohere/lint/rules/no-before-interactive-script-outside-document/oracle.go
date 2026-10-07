//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/next"
)

func oracleNoBeforeInteractiveScriptOutsideDocument() rule.Rule {
	return rules.NoBeforeInteractiveScriptOutsideDocument
}
func oracleNoBeforeInteractiveScriptOutsideDocumentOptions(fields []string) any { return nil }
