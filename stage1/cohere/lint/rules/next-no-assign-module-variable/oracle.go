//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/next"
)

func oracleNoAssignModuleVariable() rule.Rule                 { return rules.NoAssignModuleVariable }
func oracleNoAssignModuleVariableOptions(fields []string) any { return nil }
