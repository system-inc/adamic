//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/nexus"
)

func oracleWave14CorrectnessNoMockOnModuleNamespace() rule.Rule {
	return rules.CorrectnessNoMockOnModuleNamespace
}
func oracleWave14CorrectnessNoMockOnModuleNamespaceOptions(fields []string) any { return nil }
