//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/core"
)

func oracleWave22NoImpliedEval() rule.Rule                 { return rules.NoImpliedEval }
func oracleWave22NoImpliedEvalOptions(fields []string) any { return nil }
