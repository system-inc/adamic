// Probe unchanged stock declarations with declaration-only dependency stubs.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
	"os"
	"strings"
)

func main() {
	data, err := os.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}
	var records []map[string]any
	if err := json.Unmarshal(data, &records); err != nil {
		panic(err)
	}
	counts := map[string]int{"checked": 0, "refused": 0, "not_reached": 0}
	for _, record := range records {
		source := record["source"].(string)
		loaded, err := load.Load([]string{source})
		status, reason, guards := "not_reached", "", 0
		if err != nil {
			reason = err.Error()
			record["phase"] = "loader"
			record["stop_kind"] = "loader_diagnostic"
		} else {
			program, problem := lower.Lower(context.Background(), loaded)
			record["phase"] = "lowering"
			if problem != nil {
				reason = problem.Error()
				// A lowering refusal is a declaration-level observation, not proof the
				// particular any use was reached. Keep the distinction explicit.
				record["declaration_stop"] = reason
				if insideDeclaration(reason, source, int(record["first_line"].(float64)), int(record["last_line"].(float64))) {
					status = "refused"
					record["stop_kind"] = "refusal_inside_declaration"
				} else {
					record["stop_kind"] = "lowering_outside_declaration"
				}
			} else {
				for _, function := range program.Functions {
					if function.Name == "checked_any_use" || function.Name == "checked_any_property_receiver" {
						guards++
					}
				}
				// Success without a checked use is retained as not reached: it may only
				// declare a type or transport any unchanged.
				binding, _ := record["binding"].(string)
				attributed := 0
				if binding != "" {
					for _, text := range program.Strings {
						if strings.HasPrefix(text, "checked any: "+binding+" needs ") || strings.HasPrefix(text, "checked any: "+binding+".") || strings.HasPrefix(text, "checked any: "+binding+"[") {
							attributed++
						}
					}
				}
				record["attributed_scalar_messages"] = attributed
				if guards != 0 && attributed != 0 {
					status = "checked"
				} else {
					reason, record["stop_kind"] = unobservedUse(record)
				}
				record["emitted_c_bytes"] = len(native.C(program))
			}
		}
		record["status"], record["reason"], record["checked_use_helpers"] = status, reason, guards
		counts[status]++
	}
	output := map[string]any{"source_commit": "050880ce59e30b356b686bd3144efe24f875ebc8", "inventory_commit": "ea1b2359", "sites": 271, "method": "smallest enclosing unchanged declaration, ambient imports and captures with stock declaration-only type contracts", "classification_scope": "guards attributed to site bindings; refusals inside the selected stock span; every other record carries its exact diagnostic or outside-use requirement", "counts": counts, "records": records}
	result, _ := json.MarshalIndent(output, "", "  ")
	if err := os.WriteFile(os.Args[2], append(result, '\n'), 0600); err != nil {
		panic(err)
	}
	fmt.Printf("%v\n", counts)
}
