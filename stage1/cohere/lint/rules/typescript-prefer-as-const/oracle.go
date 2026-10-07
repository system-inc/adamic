//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/typescript"
)

func oraclePreferAsConst() rule.Rule                 { return rules.PreferAsConst }
func oraclePreferAsConstOptions(fields []string) any { return nil }
