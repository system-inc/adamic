//go:build lintoracle

package main

import (
	"encoding/json"
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/core"
	"strings"
)

func oracleComplexity() rule.Rule { return rules.Complexity }

func oracleComplexityOptions(fields []string) any {
	if len(fields) < 6 || fields[5] == "" || fields[5] == "null" {
		return rules.DefaultComplexitySettings()
	}
	raw := []byte(fields[5])
	// The shared all-rules rows carry a union of unrelated rule options.
	// Keep this rule's fields before invoking its strict production decoder.
	if len(fields) > 1 && fields[1] == "all" && strings.HasPrefix(fields[5], "{") {
		var union map[string]json.RawMessage
		if err := json.Unmarshal(raw, &union); err != nil {
			panic(err)
		}
		own := map[string]json.RawMessage{}
		for key, value := range union {
			switch strings.ToLower(key) {
			case "maximum", "max", "variant":
				own[key] = value
			}
		}
		var err error
		raw, err = json.Marshal(own)
		if err != nil {
			panic(err)
		}
	}
	// Captured RunWithOptions rows serialize the already decoded Go struct.
	// Decode that wire shape back into the same upstream type.
	if strings.Contains(string(raw), `"Maximum"`) || strings.Contains(string(raw), `"Variant"`) {
		options := rules.DefaultComplexitySettings()
		if err := json.Unmarshal(raw, &options); err != nil {
			panic(err)
		}
		return options
	}
	options, err := rules.DecodeComplexityOptions(raw)
	if err != nil {
		panic(err)
	}
	return options
}
