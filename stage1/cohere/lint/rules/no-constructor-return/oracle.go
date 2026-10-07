//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/core"
)

func oracleNoConstructorReturn() rule.Rule                 { return rules.NoConstructorReturn }
func oracleNoConstructorReturnOptions(fields []string) any { return nil }
