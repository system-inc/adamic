//go:build lintoracle

package main

import (
	"encoding/json"
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/core"
)

func oracleNoConstantCondition() rule.Rule { return rules.NoConstantCondition }
func oracleNoConstantConditionOptions(fields []string) any {
	var options rules.NoConstantConditionOptions
	if len(fields) > 5 && fields[5] != "" {
		// Capture serializes the already-decoded zero enum as "". It is a typed
		// default, not configuration to feed back through the custom decoder.
		var captured struct {
			CheckLoops json.RawMessage `json:"checkLoops"`
		}
		if err := json.Unmarshal([]byte(fields[5]), &captured); err != nil {
			panic(err)
		}
		if len(captured.CheckLoops) != 0 && string(captured.CheckLoops) != `""` {
			if err := json.Unmarshal(captured.CheckLoops, &options.CheckLoops); err != nil {
				panic(err)
			}
		}
	}
	return options
}
