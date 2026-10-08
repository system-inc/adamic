//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/next"
)

func oracleNextNoUnwantedPolyfillio() rule.Rule                 { return rules.NoUnwantedPolyfillio }
func oracleNextNoUnwantedPolyfillioOptions(fields []string) any { return nil }
