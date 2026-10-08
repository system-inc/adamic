//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/core"
)

func oracleCheckerNoNewWrappers() rule.Rule                 { return rules.NoNewWrappers }
func oracleCheckerNoNewWrappersOptions(fields []string) any { return nil }
