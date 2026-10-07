//go:build lintoracle

package main

import (
	"encoding/json"
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/typescript"
)

func oracleNoUnnecessaryBooleanLiteralCompare() rule.Rule {
	return rules.NoUnnecessaryBooleanLiteralCompare
}
func oracleNoUnnecessaryBooleanLiteralCompareOptions(fields []string) any {
	if len(fields) < 6 || fields[5] == "" || fields[5] == "null" {
		return nil
	}
	var wire map[string]json.RawMessage
	if err := json.Unmarshal([]byte(fields[5]), &wire); err != nil {
		panic(err)
	}
	if _, exists := wire["allowRuleToRunWithoutStrictNullChecksIKnowWhatIAmDoing"]; exists {
		options, err := rules.DecodeNoUnnecessaryBooleanLiteralCompareOptions([]byte(fields[5]))
		if err != nil {
			panic(err)
		}
		return options
	}
	options := rules.DefaultNoUnnecessaryBooleanLiteralCompareOptions()
	if err := json.Unmarshal([]byte(fields[5]), &options); err != nil {
		panic(err)
	}
	return options
}
