//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/core"
)

func oracleNoLoneBlocks() rule.Rule                 { return rules.NoLoneBlocks }
func oracleNoLoneBlocksOptions(fields []string) any { return nil }
