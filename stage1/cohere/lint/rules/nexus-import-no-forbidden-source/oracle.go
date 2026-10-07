//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/nexus"
)

func oracleForbiddenSource() rule.Rule                 { return rules.ImportNoForbiddenSource }
func oracleForbiddenSourceOptions(fields []string) any { return nil }
