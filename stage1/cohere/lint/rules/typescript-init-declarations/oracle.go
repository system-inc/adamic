//go:build lintoracle

package main

import (
	"encoding/json"

	"github.com/system-inc/cohere/internal/lint/rule"
	typescript "github.com/system-inc/cohere/internal/lint/rules/typescript"
)

func oracleInitDeclarations(options any) rule.Bound {
	return bindOracle(typescript.InitDeclarations, options)
}

// Captured rows carry the decoded upstream struct, including its deliberately inert zero mode.
func oracleInitDeclarationsOptions(fields []string) any {
	if fields[5] == "" || fields[5] == "null" {
		return nil
	}
	decoded := typescript.DefaultInitDeclarationsSettings()
	if err := json.Unmarshal([]byte(fields[5]), &decoded); err != nil {
		panic(err)
	}
	return decoded
}
