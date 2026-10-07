//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/base"
)

func oracleOrmColumnDeclare() rule.Rule                 { return rules.CorrectnessRequireOrmColumnDeclare }
func oracleOrmColumnDeclareOptions(fields []string) any { return nil }
