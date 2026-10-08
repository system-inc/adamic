//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/core"
)

func oracleNoLossOfPrecision() rule.Rule                 { return rules.NoLossOfPrecision }
func oracleNoLossOfPrecisionOptions(fields []string) any { return nil }
