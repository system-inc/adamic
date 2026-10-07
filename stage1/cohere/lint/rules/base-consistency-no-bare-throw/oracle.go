//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/base"
)

func oracleBareThrow() rule.Rule                 { return rules.ConsistencyNoBareThrow }
func oracleBareThrowOptions(fields []string) any { return nil }
