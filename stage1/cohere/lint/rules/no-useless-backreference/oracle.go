//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/core"
)

func oracleWave17NoUselessBackreference() rule.Rule { return rules.NoUselessBackreference }
func oracleWave17NoUselessBackreferenceOptions(fields []string) any {
	if len(fields) < 6 || fields[5] == "" || fields[5] == "null" {
		return nil
	}
	panic("no-useless-backreference has no options")
}
