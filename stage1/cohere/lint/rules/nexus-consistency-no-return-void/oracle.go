//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/nexus"
)

func oracleConsistencyNoReturnVoid() rule.Rule                 { return rules.ConsistencyNoReturnVoid }
func oracleConsistencyNoReturnVoidOptions(fields []string) any { return nil }
