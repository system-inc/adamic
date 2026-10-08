//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/core"
)

func oracleSymbolDescription() rule.Rule                 { return rules.SymbolDescription }
func oracleSymbolDescriptionOptions(fields []string) any { return nil }
