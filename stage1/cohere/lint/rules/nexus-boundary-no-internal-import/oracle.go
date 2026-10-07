//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/nexus"
)

func oracleNexusInternalImport() rule.Rule                 { return rules.BoundaryNoInternalImport }
func oracleNexusInternalImportOptions(fields []string) any { return nil }
