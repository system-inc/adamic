//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/adamic"
)

func oracleDefiniteAssignment() rule.Rule                 { return rules.NoDefiniteAssignment }
func oracleDefiniteAssignmentOptions(fields []string) any { return nil }
