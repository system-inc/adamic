//go:build lintoracle

package main

import (
	"encoding/json"
	"github.com/system-inc/cohere/internal/lint/rule"
	nexus "github.com/system-inc/cohere/internal/lint/rules/nexus"
)

func oracleNexusConsistencyNoLongLineComment(options any) rule.Bound {
	return bindOracle(nexus.ConsistencyNoLongLineComment, options)
}
func oracleNexusConsistencyNoLongLineCommentOptions(fields []string) any {
	var decoded nexus.ConsistencyNoLongLineCommentOptions
	if fields[5] != "" {
		if err := json.Unmarshal([]byte(fields[5]), &decoded); err != nil {
			panic(err)
		}
	}
	return decoded
}
