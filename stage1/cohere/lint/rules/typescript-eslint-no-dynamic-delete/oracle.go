//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/typescript"
)

func oracleNoDynamicDelete() rule.Rule                 { return rules.NoDynamicDelete }
func oracleNoDynamicDeleteOptions(fields []string) any { return nil }
