//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/structure"
)

func oracleStructureNetworkNoForbiddenImport() rule.Rule                 { return rules.NetworkNoForbiddenImport }
func oracleStructureNetworkNoForbiddenImportOptions(fields []string) any { return nil }
