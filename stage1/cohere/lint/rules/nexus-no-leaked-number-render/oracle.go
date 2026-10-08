//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/nexus"
)

func oracleWave14CorrectnessNoLeakedNumberRender() rule.Rule {
	return rules.CorrectnessNoLeakedNumberRender
}
func oracleWave14CorrectnessNoLeakedNumberRenderOptions(fields []string) any { return nil }
