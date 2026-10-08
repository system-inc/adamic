//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/core"
)

func oracleNoNonoctalDecimalEscape() rule.Rule                 { return rules.NoNonoctalDecimalEscape }
func oracleNoNonoctalDecimalEscapeOptions(fields []string) any { return nil }
