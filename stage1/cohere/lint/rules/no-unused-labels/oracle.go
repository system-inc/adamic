//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/core"
)

func oracleNoUnusedLabels() rule.Rule                 { return rules.NoUnusedLabels }
func oracleNoUnusedLabelsOptions(fields []string) any { return nil }
