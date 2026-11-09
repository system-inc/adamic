//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/core"
)

func oracleNoUnusedLabels(options any) rule.Bound     { return bindOracle(rules.NoUnusedLabels, options) }
func oracleNoUnusedLabelsOptions(fields []string) any { return nil }
