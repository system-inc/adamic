//go:build lintoracle

package main

import (
	"encoding/json"
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/core"
)

func oracleMaxClassesPerFile() rule.Rule { return rules.MaxClassesPerFile }

func oracleMaxClassesPerFileOptions(fields []string) any {
	decoded := rules.DefaultMaxClassesPerFileSettings()
	if len(fields) <= 5 || fields[5] == "" || fields[5] == "null" {
		return decoded
	}
	// All-rule rows can carry another rule's fields. Like the other adapters,
	// decode only our fields, retaining typed defaults rather than rejecting theirs.
	var object map[string]json.RawMessage
	if json.Unmarshal([]byte(fields[5]), &object) == nil {
		if err := json.Unmarshal([]byte(fields[5]), &decoded); err != nil {
			panic(err)
		}
		// Raw configuration calls this max; captured upstream settings call it Maximum.
		if raw, exists := object["max"]; exists {
			if err := json.Unmarshal(raw, &decoded.Maximum); err != nil {
				panic(err)
			}
		}
		if decoded.Maximum != nil && *decoded.Maximum < 1 {
			panic("max-classes-per-file takes a maximum of at least 1")
		}
		return decoded
	}
	// Raw witness/configuration options use the upstream polymorphic decoder.
	options, err := rules.DecodeMaxClassesPerFileOptions([]byte(fields[5]))
	if err != nil {
		panic(err)
	}
	return options
}
