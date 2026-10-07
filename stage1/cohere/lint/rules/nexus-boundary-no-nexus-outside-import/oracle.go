//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/nexus"
)

func oracleNexusOutsideImport() rule.Rule                 { return rules.BoundaryNoNexusOutsideImport }
func oracleNexusOutsideImportOptions(fields []string) any { return nil }
