//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/structure"
)

func oracleStructureNextNoPageState() rule.Rule                 { return rules.NextNoPageState }
func oracleStructureNextNoPageStateOptions(fields []string) any { return nil }
