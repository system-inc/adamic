//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/typescript"
)

func oracleNoConfusingNonNullAssertion() rule.Rule                 { return rules.NoConfusingNonNullAssertion }
func oracleNoConfusingNonNullAssertionOptions(fields []string) any { return nil }
