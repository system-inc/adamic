//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/typescript"
)

func oracleNoImportTypeSideEffects() rule.Rule                 { return rules.NoImportTypeSideEffects }
func oracleNoImportTypeSideEffectsOptions(fields []string) any { return nil }
