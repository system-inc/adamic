//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/core"
)

func oracleNoPrototypeBuiltins() rule.Rule                 { return rules.NoPrototypeBuiltins }
func oracleNoPrototypeBuiltinsOptions(fields []string) any { return nil }
