//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/core"
)

func oracleDotNotation() rule.Rule { return rules.DotNotation }
func oracleDotNotationOptions(fields []string) any {
	// The all row is unconfigured for directory-owned rules. Its legacy option bag
	// configures the baseline adapters; named rows carry this rule's settings.
	if len(fields) > 1 && fields[1] == "all" {
		return nil
	}
	if len(fields) < 6 || fields[5] == "" || fields[5] == "null" {
		return nil
	}
	options, err := rules.DecodeDotNotationOptions([]byte(fields[5]))
	if err != nil {
		panic(err)
	}
	return options
}
