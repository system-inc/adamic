//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/next"
)

func oracleModuleVariable() rule.Rule                 { return rules.NoAssignModuleVariable }
func oracleModuleVariableOptions(fields []string) any { return nil }
