//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/core"
)

func oracleWave22PreferObjectHasOwn() rule.Rule                 { return rules.PreferObjectHasOwn }
func oracleWave22PreferObjectHasOwnOptions(fields []string) any { return nil }
