//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/next"
)

func oracleNextNoTypos() rule.Rule                 { return rules.NoTypos }
func oracleNextNoTyposOptions(fields []string) any { return nil }
