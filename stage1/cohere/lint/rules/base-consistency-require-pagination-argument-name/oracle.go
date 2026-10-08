//go:build lintoracle

package main

import (
	"encoding/json"
	"github.com/system-inc/cohere/internal/lint/rule"
	base "github.com/system-inc/cohere/internal/lint/rules/base"
)

func oracleRequirePaginationArgumentName() rule.Rule {
	return base.ConsistencyRequirePaginationArgumentName
}

// Upstream has no options type and ignores its options argument. Decode and retain
// the actual JSON payload so the shared no-dropped-options guard remains intact.
func oracleRequirePaginationArgumentNameOptions(fields []string) any {
	var options any = struct{}{}
	if len(fields) > 5 && fields[5] != "" && fields[5] != "null" {
		if err := json.Unmarshal([]byte(fields[5]), &options); err != nil {
			panic(err)
		}
	}
	return options
}
