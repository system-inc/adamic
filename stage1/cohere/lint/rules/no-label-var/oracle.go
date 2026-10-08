//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/core"
)

func oracleWave14NoLabelVar() rule.Rule                 { return rules.NoLabelVar }
func oracleWave14NoLabelVarOptions(fields []string) any { return nil }
