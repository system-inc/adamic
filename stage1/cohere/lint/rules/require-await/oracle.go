//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/core"
)

func oracleRequireAwait() rule.Rule                 { return rules.RequireAwait }
func oracleRequireAwaitOptions(fields []string) any { return nil }
