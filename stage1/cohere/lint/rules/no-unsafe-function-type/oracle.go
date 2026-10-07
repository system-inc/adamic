//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/typescript"
)

func oracleNoUnsafeFunctionType() rule.Rule                 { return rules.NoUnsafeFunctionType }
func oracleNoUnsafeFunctionTypeOptions(fields []string) any { return nil }
