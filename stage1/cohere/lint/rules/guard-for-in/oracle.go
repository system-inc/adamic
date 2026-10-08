//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/core"
)

func oracleGuardForIn() rule.Rule                 { return rules.GuardForIn }
func oracleGuardForInOptions(fields []string) any { return nil }
