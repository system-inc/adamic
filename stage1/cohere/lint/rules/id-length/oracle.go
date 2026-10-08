//go:build lintoracle

package main

import (
	"encoding/json"
	esregexp "github.com/system-inc/cohere/internal/lint/ecmascript/regexp"
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/core"
)

func oracleIdentifierLength() rule.Rule { return rules.IdLength }
func oracleIdentifierLengthOptions(fields []string) any {
	// The all row is unconfigured for directory-owned rules. Its legacy option bag
	// configures the baseline adapters; named rows carry this rule's settings.
	if len(fields) > 1 && fields[1] == "all" {
		return nil
	}
	if len(fields) < 6 || fields[5] == "" || fields[5] == "null" {
		return nil
	}
	var object map[string]json.RawMessage
	if err := json.Unmarshal([]byte(fields[5]), &object); err != nil {
		panic(err)
	}
	if _, ok := object["Minimum"]; !ok {
		options, err := rules.DecodeIdLengthOptions([]byte(fields[5]))
		if err != nil {
			panic(err)
		}
		return options
	}
	var patterns []json.RawMessage
	if raw, ok := object["ExceptionPatterns"]; ok {
		if err := json.Unmarshal(raw, &patterns); err != nil {
			panic(err)
		}
	}
	delete(object, "ExceptionPatterns")
	encoded, err := json.Marshal(object)
	if err != nil {
		panic(err)
	}
	var options rules.IdLengthSettings
	if err := json.Unmarshal(encoded, &options); err != nil {
		panic(err)
	}
	for _, raw := range patterns {
		var source string
		if err := json.Unmarshal(raw, &source); err != nil {
			panic("id-length: shared upstream capture omitted esregexp.Source from ExceptionPatterns")
		}
		pattern, err := esregexp.Compile(source, "u")
		if err != nil {
			panic(err)
		}
		options.ExceptionPatterns = append(options.ExceptionPatterns, pattern)
	}
	return options
}
