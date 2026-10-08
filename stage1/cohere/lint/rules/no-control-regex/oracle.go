//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/core"
)

func oracleNoControlRegex() rule.Rule                 { return rules.NoControlRegex }
func oracleNoControlRegexOptions(fields []string) any { return nil }
