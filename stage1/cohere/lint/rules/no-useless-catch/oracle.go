//go:build lintoracle

package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/core"
)

func oracleNoUselessCatch() rule.Rule { return rules.NoUselessCatch }

// Upstream declares no options type and ignores options. Decode supplied JSON
// without inventing fields or returning nil for a row that carries options.
func oracleNoUselessCatchOptions(fields []string) any {
	if len(fields) <= 5 || strings.TrimSpace(fields[5]) == "" {
		return nil
	}
	var options any
	if err := json.Unmarshal([]byte(fields[5]), &options); err != nil {
		panic(fmt.Sprintf("no-useless-catch: options: %v", err))
	}
	if options == nil {
		return struct{}{}
	}
	return options
}
