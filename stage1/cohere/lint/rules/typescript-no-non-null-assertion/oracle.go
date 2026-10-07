//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/typescript"
)

func oracleNonNullAssertion() rule.Rule                 { return rules.NoNonNullAssertion }
func oracleNonNullAssertionOptions(fields []string) any { return nil }
