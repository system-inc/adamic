//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/structure"
)

func oracleOrganizedImports() rule.Rule                 { return rules.ConsistencyRequireOrganizedImports }
func oracleOrganizedImportsOptions(fields []string) any { return nil }
