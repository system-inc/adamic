//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/typescript"
)

func oracleNoNonNullAssertion() rule.Rule                 { return rules.NoNonNullAssertion }
func oracleNoNonNullAssertionOptions(fields []string) any { return nil }
