//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/nexus"
)

func oracleWave22ConcurrencyNoLostUpdate() rule.Rule                 { return rules.ConcurrencyNoLostUpdate }
func oracleWave22ConcurrencyNoLostUpdateOptions(fields []string) any { return nil }
