//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/typescript"
)

func oraclePreferNamespaceKeyword() rule.Rule                 { return rules.PreferNamespaceKeyword }
func oraclePreferNamespaceKeywordOptions(fields []string) any { return nil }
