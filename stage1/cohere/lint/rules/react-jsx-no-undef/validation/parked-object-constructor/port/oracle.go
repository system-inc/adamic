//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/core"
)

func oracleWave21ObjectConstructor() rule.Rule                 { return rules.NoObjectConstructor }
func oracleWave21ObjectConstructorOptions(fields []string) any { return nil }
