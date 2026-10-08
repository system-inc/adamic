//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/typescript"
)

func oracleNoDuplicateEnumValues() rule.Rule                 { return rules.NoDuplicateEnumValues }
func oracleNoDuplicateEnumValuesOptions(fields []string) any { return nil }
