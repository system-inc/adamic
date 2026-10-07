//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/nexus"
)

func oracleImportRequireNodeNamespace() rule.Rule                 { return rules.ImportRequireNodeNamespace }
func oracleImportRequireNodeNamespaceOptions(fields []string) any { return nil }
