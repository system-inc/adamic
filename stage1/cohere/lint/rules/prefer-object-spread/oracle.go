//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/core"
)

func oracleWave22PreferObjectSpread() rule.Rule                 { return rules.PreferObjectSpread }
func oracleWave22PreferObjectSpreadOptions(fields []string) any { return nil }
