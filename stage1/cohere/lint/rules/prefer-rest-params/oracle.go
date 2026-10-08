//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/core"
)

func oraclePreferRestParams() rule.Rule                 { return rules.PreferRestParams }
func oraclePreferRestParamsOptions(fields []string) any { return nil }
