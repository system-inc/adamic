//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/structure"
)

func oracleStructureNextRequireApiParameterName() rule.Rule                 { return rules.NextRequireApiParameterName }
func oracleStructureNextRequireApiParameterNameOptions(fields []string) any { return nil }
