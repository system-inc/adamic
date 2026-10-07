//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/core"
)

func oracleForDirectionCore() rule.Rule                 { return rules.ForDirection }
func oracleForDirectionCoreOptions(fields []string) any { return nil }
