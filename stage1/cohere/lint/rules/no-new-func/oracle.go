//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/core"
)

func oracleWave06NoNewFunc() rule.Rule                 { return rules.NoNewFunc }
func oracleWave06NoNewFuncOptions(fields []string) any { return nil }
