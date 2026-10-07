//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/next"
)

func oracleNextDocumentImport() rule.Rule                 { return rules.NoDocumentImportInPage }
func oracleNextDocumentImportOptions(fields []string) any { return nil }
