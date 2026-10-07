//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/core"
)

func oracleNoDeleteVar() rule.Rule                 { return rules.NoDeleteVar }
func oracleNoDeleteVarOptions(fields []string) any { return nil }
