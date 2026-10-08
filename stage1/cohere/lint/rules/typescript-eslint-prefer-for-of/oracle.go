//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/typescript"
)

func oraclePreferForOf() rule.Rule                 { return rules.PreferForOf }
func oraclePreferForOfOptions(fields []string) any { return nil }
