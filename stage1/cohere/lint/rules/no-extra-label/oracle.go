//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/core"
)

func oracleNoExtraLabel() rule.Rule                 { return rules.NoExtraLabel }
func oracleNoExtraLabelOptions(fields []string) any { return nil }
