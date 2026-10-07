//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/typescript"
)

func oracleEnumInitializers() rule.Rule                 { return rules.PreferEnumInitializers }
func oracleEnumInitializersOptions(fields []string) any { return nil }
