//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/next"
)

func oracleNextNoHeadElement() rule.Rule                 { return rules.NoHeadElement }
func oracleNextNoHeadElementOptions(fields []string) any { return nil }
