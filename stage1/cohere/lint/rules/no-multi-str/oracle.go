//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/core"
)

func oracleNoMultiStr() rule.Rule                 { return rules.NoMultiStr }
func oracleNoMultiStrOptions(fields []string) any { return nil }
