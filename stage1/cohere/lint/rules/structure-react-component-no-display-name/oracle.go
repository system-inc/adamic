//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/structure"
)

func oracleStructureReactComponentNoDisplayName() rule.Rule                 { return rules.ReactComponentNoDisplayName }
func oracleStructureReactComponentNoDisplayNameOptions(fields []string) any { return nil }
