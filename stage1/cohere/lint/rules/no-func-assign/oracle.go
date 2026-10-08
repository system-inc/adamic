//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/core"
)

func oracleNoFuncAssign() rule.Rule                 { return rules.NoFuncAssign }
func oracleNoFuncAssignOptions(fields []string) any { return nil }
