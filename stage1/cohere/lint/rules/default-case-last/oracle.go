//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/core"
)

func oracleDefaultCaseLastCore() rule.Rule                 { return rules.DefaultCaseLast }
func oracleDefaultCaseLastCoreOptions(fields []string) any { return nil }
