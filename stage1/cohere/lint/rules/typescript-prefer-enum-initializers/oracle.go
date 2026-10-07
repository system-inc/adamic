//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/typescript"
)

func oraclePreferEnumInitializers() rule.Rule                 { return rules.PreferEnumInitializers }
func oraclePreferEnumInitializersOptions(fields []string) any { return nil }
