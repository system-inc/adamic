//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/typescript"
)

func oracleWave14NoArrayDelete() rule.Rule                 { return rules.NoArrayDelete }
func oracleWave14NoArrayDeleteOptions(fields []string) any { return nil }
