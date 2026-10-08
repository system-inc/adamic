//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/react"
)

func oracleNoArrowFunctionLifecycle() rule.Rule                 { return rules.NoArrowFunctionLifecycle }
func oracleNoArrowFunctionLifecycleOptions(fields []string) any { return nil }
