//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/core"
)

func oracleNoCaseDeclarations() rule.Rule                 { return rules.NoCaseDeclarations }
func oracleNoCaseDeclarationsOptions(fields []string) any { return nil }
