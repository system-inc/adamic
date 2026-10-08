//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/structure"
)

func oracleNetworkNoInvalidateCacheLiteralKey() rule.Rule {
	return rules.NetworkNoInvalidateCacheLiteralKey
}
func oracleNetworkNoInvalidateCacheLiteralKeyOptions(fields []string) any { return nil }
