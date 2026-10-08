//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/core"
)

func oracleNoEqNull() rule.Rule                 { return rules.NoEqNull }
func oracleNoEqNullOptions(fields []string) any { return nil }
