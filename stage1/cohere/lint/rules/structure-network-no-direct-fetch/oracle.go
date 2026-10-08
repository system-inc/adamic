//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/structure"
)

func oracleStructureNetworkNoDirectFetch() rule.Rule                 { return rules.NetworkNoDirectFetch }
func oracleStructureNetworkNoDirectFetchOptions(fields []string) any { return nil }
