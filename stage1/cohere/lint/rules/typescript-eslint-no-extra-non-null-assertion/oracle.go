//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/typescript"
)

func oracleNoExtraNonNullAssertion() rule.Rule                 { return rules.NoExtraNonNullAssertion }
func oracleNoExtraNonNullAssertionOptions(fields []string) any { return nil }
