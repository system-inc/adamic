//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/core"
)

func oracleNoCompareNegZero() rule.Rule                 { return rules.NoCompareNegZero }
func oracleNoCompareNegZeroOptions(fields []string) any { return nil }
