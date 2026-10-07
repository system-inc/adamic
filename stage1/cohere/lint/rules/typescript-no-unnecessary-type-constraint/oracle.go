//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/typescript"
)

func oracleUnnecessaryConstraint() rule.Rule                 { return rules.NoUnnecessaryTypeConstraint }
func oracleUnnecessaryConstraintOptions(fields []string) any { return nil }
