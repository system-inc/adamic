//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/nexus"
)

func oracleWave17CorrectnessNoCollectionMisuse() rule.Rule {
	return rules.CorrectnessNoCollectionMisuse
}
func oracleWave17CorrectnessNoCollectionMisuseOptions(fields []string) any {
	if len(fields) < 6 || fields[5] == "" || fields[5] == "null" {
		return nil
	}
	panic("nexus/correctness-no-collection-misuse has no options")
}
