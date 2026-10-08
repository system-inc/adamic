//go:build lintoracle

package main

import (
	"encoding/json"

	"github.com/system-inc/cohere/internal/lint/rule"
	typescript "github.com/system-inc/cohere/internal/lint/rules/typescript"
)

func oracleConsistentTypeDefinitions() rule.Rule { return typescript.ConsistentTypeDefinitions }

// Captured options are the decoded struct; witness options can use the upstream string schema.
func oracleConsistentTypeDefinitionsOptions(fields []string) any {
	settings := typescript.DefaultConsistentTypeDefinitionsSettings()
	if len(fields) <= 5 || fields[5] == "" || fields[5] == "null" {
		return settings
	}
	raw := []byte(fields[5])
	if raw[0] == '"' {
		decoded, err := typescript.DecodeConsistentTypeDefinitionsOptions(raw)
		if err != nil {
			panic(err)
		}
		return decoded
	}
	if err := json.Unmarshal(raw, &settings); err != nil {
		panic(err)
	}
	if settings.Style != typescript.ConsistentTypeDefinitionsInterface && settings.Style != typescript.ConsistentTypeDefinitionsType {
		panic("consistent-type-definitions: invalid style")
	}
	return settings
}
