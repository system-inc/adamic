//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/typescript"
)

func oracleNoUnnecessaryTypeConstraint() rule.Rule                 { return rules.NoUnnecessaryTypeConstraint }
func oracleNoUnnecessaryTypeConstraintOptions(fields []string) any { return nil }
