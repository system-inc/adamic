//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/structure"
)

func oracleStructureReactComponentNoForwardRef() rule.Rule                 { return rules.ReactComponentNoForwardRef }
func oracleStructureReactComponentNoForwardRefOptions(fields []string) any { return nil }
