//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/structure"
)

func oracleNetworkNoStringLiteralQuery() rule.Rule                 { return rules.NetworkNoStringLiteralQuery }
func oracleNetworkNoStringLiteralQueryOptions(fields []string) any { return nil }
