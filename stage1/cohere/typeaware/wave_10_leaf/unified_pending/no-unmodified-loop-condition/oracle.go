//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/core"
)

func oracleNoUnmodifiedLoopCondition() rule.Rule                 { return rules.NoUnmodifiedLoopCondition }
func oracleNoUnmodifiedLoopConditionOptions(fields []string) any { return nil }
