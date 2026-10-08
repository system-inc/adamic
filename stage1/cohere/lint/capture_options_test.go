package lint

import (
	"encoding/json"
	"fmt"
	"testing"
)

// Captures serialize decoded Go values, not user configuration. This upstream
// entry type has no marshal tags, but its strict decoder requires lower-case
// schema keys. Rename those two fields only; preserve entries, order and values.
func capturedSchemaOptions(name string, raw json.RawMessage) (json.RawMessage, error) {
	if name != "preserve-caught-error" || len(raw) == 0 || string(raw) == "null" {
		return raw, nil
	}
	var options map[string]json.RawMessage
	if err := json.Unmarshal(raw, &options); err != nil {
		return nil, err
	}
	classes, exists := options["errorClassNames"]
	if !exists || string(classes) == "null" {
		return raw, nil
	}
	var entries []json.RawMessage
	if err := json.Unmarshal(classes, &entries); err != nil {
		return nil, err
	}
	for i, entry := range entries {
		if len(entry) == 0 || entry[0] != '{' {
			continue
		}
		var fields map[string]json.RawMessage
		if err := json.Unmarshal(entry, &fields); err != nil {
			return nil, err
		}
		for old, next := range map[string]string{"Name": "name", "ArgumentPosition": "argumentPosition"} {
			if value, exists := fields[old]; exists {
				if _, conflict := fields[next]; conflict {
					return nil, fmt.Errorf("captured %s has both %s and %s", name, old, next)
				}
				fields[next] = value
				delete(fields, old)
			}
		}
		var err error
		entries[i], err = json.Marshal(fields)
		if err != nil {
			return nil, err
		}
	}
	var err error
	options["errorClassNames"], err = json.Marshal(entries)
	if err != nil {
		return nil, err
	}
	return json.Marshal(options)
}

func TestCapturedCaughtErrorOptions(t *testing.T) {
	raw := json.RawMessage(`{"requireCatchParameter":true,"errorClassNames":[{"Name":"AppError","ArgumentPosition":2},"Other",{"name":"AppError","argumentPosition":3}]}`)
	got, err := capturedSchemaOptions("preserve-caught-error", raw)
	if err != nil {
		t.Fatal(err)
	}
	want := `{"errorClassNames":[{"argumentPosition":2,"name":"AppError"},"Other",{"argumentPosition":3,"name":"AppError"}],"requireCatchParameter":true}`
	if string(got) != want {
		t.Fatalf("captured options lost a value or order: %s", got)
	}
	unchanged, err := capturedSchemaOptions("other-rule", raw)
	if err != nil || string(unchanged) != string(raw) {
		t.Fatal("unrelated options changed")
	}
	if _, err := capturedSchemaOptions("preserve-caught-error", json.RawMessage(`{"errorClassNames":[{"Name":"AppError","name":"Other"}]}`)); err == nil {
		t.Fatal("conflicting captured fields were accepted")
	}
}
