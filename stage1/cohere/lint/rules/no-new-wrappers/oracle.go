//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/core"
)

func oracleNoNewWrappers() rule.Rule                 { return rules.NoNewWrappers }
func oracleNoNewWrappersOptions(fields []string) any { return nil }
