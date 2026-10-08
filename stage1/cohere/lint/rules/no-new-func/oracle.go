//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/core"
)

func oracleNoNewFunc() rule.Rule                 { return rules.NoNewFunc }
func oracleNoNewFuncOptions(fields []string) any { return nil }
