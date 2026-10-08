//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/next"
)

func oracleNextNoHeadImportInDocument() rule.Rule                 { return rules.NoHeadImportInDocument }
func oracleNextNoHeadImportInDocumentOptions(fields []string) any { return nil }
