//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/typescript"
)

func oracleNoRedundantTypeConstituents() rule.Rule                 { return rules.NoRedundantTypeConstituents }
func oracleNoRedundantTypeConstituentsOptions(fields []string) any { return nil }
