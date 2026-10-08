//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/core"
)

func oracleNoThrowLiteral() rule.Rule                 { return rules.NoThrowLiteral }
func oracleNoThrowLiteralOptions(fields []string) any { return nil }
