//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/nexus"
)

func oracleCheckerIsoStringDateCut() rule.Rule                 { return rules.ConsistencyNoIsoStringDateCut }
func oracleCheckerIsoStringDateCutOptions(fields []string) any { return nil }
