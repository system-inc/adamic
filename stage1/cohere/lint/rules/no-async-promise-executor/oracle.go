//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/core"
)

func oracleNoAsyncPromiseExecutor() rule.Rule                 { return rules.NoAsyncPromiseExecutor }
func oracleNoAsyncPromiseExecutorOptions(fields []string) any { return nil }
