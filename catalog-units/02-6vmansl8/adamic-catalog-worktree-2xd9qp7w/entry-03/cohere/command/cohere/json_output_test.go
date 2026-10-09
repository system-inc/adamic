package main

import (
	"bytes"
	"encoding/json"
	"os"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"
)

func TestJSONLines(t *testing.T) {
	t.Parallel()
	var out bytes.Buffer
	finding := runFinding{Path: "/repo/a.ts", Line: 3, Column: 7, Severity: "error", Rule: "prefer-const", MessageID: "useConst", Message: "Use const."}
	if err := writeJSONLine(&out, findingAsJSON(finding)); err != nil {
		t.Fatal(err)
	}
	for _, line := range changedFilesAsJSON([]changedFile{
		{Path: "b.ts", Fixed: true, Formatted: true, FixedBy: map[string]int{"prefer-const": 2, "object-shorthand": 1}},
	}) {
		if err := writeJSONLine(&out, line); err != nil {
			t.Fatal(err)
		}
	}
	want := `{"kind":"finding","path":"/repo/a.ts","line":3,"column":7,"severity":"error","rule":"prefer-const","messageId":"useConst","message":"Use const."}
{"kind":"fixed","path":"b.ts","rules":["object-shorthand","prefer-const"]}
{"kind":"formatted","path":"b.ts"}
`
	if out.String() != want {
		t.Errorf("lines:\n got\n%s want\n%s", out.String(), want)
	}
}

func TestJSONSummary(t *testing.T) {
	t.Parallel()
	summary := cleanSummary()
	warmSummary(&summary)
	summary.Findings = 2
	summary.Changed = []changedFile{{Path: "b.ts", Fixed: true, Formatted: true}, {Path: "c.ts", Formatted: true}}
	summary.Gaps.CrashedFiles = 1

	encoded, err := json.Marshal(summaryAsJSON(summary))
	if err != nil {
		t.Fatal(err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatal(err)
	}
	for key, want := range map[string]any{
		"kind": "summary", "schemaVersion": float64(outputSchemaVersion), "verdict": "fail", "seconds": 0.7,
		"findings": float64(2), "filesFixed": float64(1), "filesFormatted": float64(2),
		"cohered": float64(2), "checked": float64(3), "cached": float64(3923), "filesInScope": float64(3926), "rules": float64(480), "formattingSeconds": 0.02,
	} {
		if decoded[key] != want {
			t.Errorf("summary %s = %v, want %v", key, decoded[key], want)
		}
	}
	if gaps := decoded["gaps"].(map[string]any); gaps["crashedFiles"] != float64(1) {
		t.Errorf("the summary's gaps dropped the crashed file: %v", gaps)
	}
	if phases := decoded["phases"].([]any); len(phases) != 3 || phases[0].(map[string]any)["name"] != "fix" {
		t.Errorf("phases: %v", phases)
	}
	if replay := summaryAsJSON(runSummary{Total: 50 * time.Millisecond, Cache: cacheUse{Replayed: true}}); !replay.Cache.Replayed || replay.Verdict != "pass" {
		t.Errorf("a replay's summary: %+v", replay)
	}
}

// TestOutputSchemaMatchesTheTypes holds schema/CohereOutput.schema.json to the types `--json` writes:
// every field a type writes is a property in the schema, every property is a field, and the required
// list is exactly the fields that are never omitted. A field added to a type without the schema, or a
// property the code stopped writing, fails here rather than in a consumer.
func TestOutputSchemaMatchesTheTypes(t *testing.T) {
	t.Parallel()
	raw, err := os.ReadFile("../../schema/CohereOutput.schema.json")
	if err != nil {
		t.Fatal(err)
	}
	var schema struct {
		Defs map[string]schemaObject `json:"$defs"`
	}
	if err := json.Unmarshal(raw, &schema); err != nil {
		t.Fatal(err)
	}
	summary := schema.Defs["summary"]
	for name, check := range map[string]struct {
		object schemaObject
		value  any
	}{
		"finding":       {schema.Defs["finding"], findingJSON{}},
		"changedFile":   {schema.Defs["changedFile"], changedFileJSON{}},
		"summary":       {summary, summaryJSON{}},
		"summary.phase": {*summary.Properties["phases"].Items, phaseJSON{}},
		"summary.cache": {summary.Properties["cache"], cacheUse{}},
		"summary.gaps":  {summary.Properties["gaps"], runGaps{}},
	} {
		fields, required := jsonFields(reflect.TypeOf(check.value))
		if got := schemaPropertyNames(check.object.Properties); strings.Join(got, ",") != strings.Join(fields, ",") {
			t.Errorf("%s: the schema's properties are %v, the type writes %v", name, got, fields)
		}
		schemaRequired := append([]string(nil), check.object.Required...)
		sort.Strings(schemaRequired)
		if strings.Join(schemaRequired, ",") != strings.Join(required, ",") {
			t.Errorf("%s: the schema requires %v, the type always writes %v", name, schemaRequired, required)
		}
	}
}

type schemaObject struct {
	Properties map[string]schemaObject `json:"properties"`
	Required   []string                `json:"required"`
	Items      *schemaObject           `json:"items"`
}

// jsonFields is the names a type writes, sorted, and those it never omits, with embedded structs' fields
// flattened in as encoding/json flattens them.
func jsonFields(structType reflect.Type) (fields, required []string) {
	for index := range structType.NumField() {
		field := structType.Field(index)
		if field.Anonymous {
			embeddedFields, embeddedRequired := jsonFields(field.Type)
			fields, required = append(fields, embeddedFields...), append(required, embeddedRequired...)
			continue
		}
		name, options, _ := strings.Cut(field.Tag.Get("json"), ",")
		if name == "" || name == "-" {
			continue
		}
		fields = append(fields, name)
		if !strings.Contains(options, "omitempty") {
			required = append(required, name)
		}
	}
	sort.Strings(fields)
	sort.Strings(required)
	return fields, required
}

func schemaPropertyNames(properties map[string]schemaObject) []string {
	keys := make([]string, 0, len(properties))
	for key := range properties {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
