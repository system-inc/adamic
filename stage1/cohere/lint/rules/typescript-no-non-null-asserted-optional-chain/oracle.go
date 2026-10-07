//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/typescript"
)

func oracleOptionalAssertion() rule.Rule                 { return rules.NoNonNullAssertedOptionalChain }
func oracleOptionalAssertionOptions(fields []string) any { return nil }
