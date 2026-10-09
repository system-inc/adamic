//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/structure"
)

func oracleStructureImportRequireReactNamespace() rule.Rule                 { return rules.ImportRequireReactNamespace }
func oracleStructureImportRequireReactNamespaceOptions(fields []string) any { return nil }
