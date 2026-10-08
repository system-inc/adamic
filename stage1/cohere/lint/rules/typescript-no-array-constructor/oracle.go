//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/core"
)

func oracleFactsWave19NoArrayConstructor() rule.Rule                 { return rules.NoArrayConstructor }
func oracleFactsWave19NoArrayConstructorOptions(fields []string) any { return nil }
