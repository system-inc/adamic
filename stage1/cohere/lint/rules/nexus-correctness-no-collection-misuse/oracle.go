//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/nexus"
)

func oracleCorrectnessNoCollectionMisuse() rule.Rule                 { return rules.CorrectnessNoCollectionMisuse }
func oracleCorrectnessNoCollectionMisuseOptions(fields []string) any { return nil }
