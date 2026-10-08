//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/core"
)

func oracleCheckerNoClassAssign() rule.Rule                 { return rules.NoClassAssign }
func oracleCheckerNoClassAssignOptions(fields []string) any { return nil }
