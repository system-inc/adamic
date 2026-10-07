//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/next"
)

func oracleNoDocumentImportInPage() rule.Rule                 { return rules.NoDocumentImportInPage }
func oracleNoDocumentImportInPageOptions(fields []string) any { return nil }
