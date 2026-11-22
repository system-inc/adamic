package oracle

import (
	"context"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"path/filepath"
	"testing"
)

// Erasing the RegExp type before the cast must not expose a prototype value,
// a wrong presence answer or a receiverless call. Existing cast and opaque-host
// rules stop these roads before any backend runs, on main and the served head.
var regexpErasedCastRows = []struct{ name, path, stdout, diagnostic string }{
	{"unknown_value_as_a", "internal/oracle/testdata/regexp_served_casts/unknown_value_as.a", "function\n", ":3:11: Adamic 0.1 refuses a cast the runtime can't check; use a proven upcast, narrow a discriminated union by its tag (`if (shape.kind !== 'Circle') panic('why')`), downcast along nominal class ancestry, or move the file to `.ts`, where a member cast is checked (adamic/no-unchecked-cast)"},
	{"unknown_value_as_ts", "internal/oracle/testdata/regexp_served_casts/unknown_value_as.ts", "function\n", ":3:11: Adamic 0.1 refuses a cast the runtime can't check: a cast from unknown to Maybe; narrow it first with a type predicate or a typeof and in test, then cast from what they prove (adamic/no-unchecked-cast)"},
	{"unknown_presence_as_a", "internal/oracle/testdata/regexp_served_casts/unknown_presence_as.a", "false\n", ":3:11: Adamic 0.1 refuses a cast the runtime can't check; use a proven upcast, narrow a discriminated union by its tag (`if (shape.kind !== 'Circle') panic('why')`), downcast along nominal class ancestry, or move the file to `.ts`, where a member cast is checked (adamic/no-unchecked-cast)"},
	{"unknown_presence_as_ts", "internal/oracle/testdata/regexp_served_casts/unknown_presence_as.ts", "false\n", ":3:11: Adamic 0.1 refuses a cast the runtime can't check: a cast from unknown to Maybe; narrow it first with a type predicate or a typeof and in test, then cast from what they prove (adamic/no-unchecked-cast)"},
	{"unknown_call_as_a", "internal/oracle/testdata/regexp_served_casts/unknown_call_as.a", "true\n", ":3:11: Adamic 0.1 refuses a cast the runtime can't check; use a proven upcast, narrow a discriminated union by its tag (`if (shape.kind !== 'Circle') panic('why')`), downcast along nominal class ancestry, or move the file to `.ts`, where a member cast is checked (adamic/no-unchecked-cast)"},
	{"unknown_call_as_ts", "internal/oracle/testdata/regexp_served_casts/unknown_call_as.ts", "true\n", ":3:11: Adamic 0.1 refuses a cast the runtime can't check: a cast from unknown to Maybe; narrow it first with a type predicate or a typeof and in test, then cast from what they prove (adamic/no-unchecked-cast)"},
	{"field_value_as_a", "internal/oracle/testdata/regexp_served_casts/field_value_as.a", "function\n", ":3:11: Adamic 0.1 refuses optional property test in Maybe absent from structural source object, which can hide fields; declare test on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"},
	{"field_value_as_ts", "internal/oracle/testdata/regexp_served_casts/field_value_as.ts", "function\n", ":3:11: Adamic 0.1 refuses optional property test in Maybe absent from structural source object, which can hide fields; declare test on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"},
	{"field_presence_as_a", "internal/oracle/testdata/regexp_served_casts/field_presence_as.a", "false\n", ":3:11: Adamic 0.1 refuses optional property test in Maybe absent from structural source object, which can hide fields; declare test on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"},
	{"field_presence_as_ts", "internal/oracle/testdata/regexp_served_casts/field_presence_as.ts", "false\n", ":3:11: Adamic 0.1 refuses optional property test in Maybe absent from structural source object, which can hide fields; declare test on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"},
	{"field_call_as_a", "internal/oracle/testdata/regexp_served_casts/field_call_as.a", "true\n", ":3:11: Adamic 0.1 refuses optional property test in Maybe absent from structural source object, which can hide fields; declare test on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"},
	{"field_call_as_ts", "internal/oracle/testdata/regexp_served_casts/field_call_as.ts", "true\n", ":3:11: Adamic 0.1 refuses optional property test in Maybe absent from structural source object, which can hide fields; declare test on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"},
	{"unknown_value_angle_a", "internal/oracle/testdata/regexp_served_casts/unknown_value_angle.a", "function\n", ":2:20: stage 0 can't lower an opaque host or collection value viewed as unknown or object (dynamic property metadata) yet"},
	{"unknown_value_angle_ts", "internal/oracle/testdata/regexp_served_casts/unknown_value_angle.ts", "function\n", ":2:20: stage 0 can't lower an opaque host or collection value viewed as unknown or object (dynamic property metadata) yet"},
	{"unknown_presence_angle_a", "internal/oracle/testdata/regexp_served_casts/unknown_presence_angle.a", "false\n", ":2:20: stage 0 can't lower an opaque host or collection value viewed as unknown or object (dynamic property metadata) yet"},
	{"unknown_presence_angle_ts", "internal/oracle/testdata/regexp_served_casts/unknown_presence_angle.ts", "false\n", ":2:20: stage 0 can't lower an opaque host or collection value viewed as unknown or object (dynamic property metadata) yet"},
	{"unknown_call_angle_a", "internal/oracle/testdata/regexp_served_casts/unknown_call_angle.a", "true\n", ":2:20: stage 0 can't lower an opaque host or collection value viewed as unknown or object (dynamic property metadata) yet"},
	{"unknown_call_angle_ts", "internal/oracle/testdata/regexp_served_casts/unknown_call_angle.ts", "true\n", ":2:20: stage 0 can't lower an opaque host or collection value viewed as unknown or object (dynamic property metadata) yet"},
	{"field_value_angle_a", "internal/oracle/testdata/regexp_served_casts/field_value_angle.a", "function\n", ":2:33: stage 0 can't lower an opaque host or collection value viewed as unknown or object (dynamic property metadata) yet"},
	{"field_value_angle_ts", "internal/oracle/testdata/regexp_served_casts/field_value_angle.ts", "function\n", ":2:33: stage 0 can't lower an opaque host or collection value viewed as unknown or object (dynamic property metadata) yet"},
	{"field_presence_angle_a", "internal/oracle/testdata/regexp_served_casts/field_presence_angle.a", "false\n", ":2:33: stage 0 can't lower an opaque host or collection value viewed as unknown or object (dynamic property metadata) yet"},
	{"field_presence_angle_ts", "internal/oracle/testdata/regexp_served_casts/field_presence_angle.ts", "false\n", ":2:33: stage 0 can't lower an opaque host or collection value viewed as unknown or object (dynamic property metadata) yet"},
	{"field_call_angle_a", "internal/oracle/testdata/regexp_served_casts/field_call_angle.a", "true\n", ":2:33: stage 0 can't lower an opaque host or collection value viewed as unknown or object (dynamic property metadata) yet"},
	{"field_call_angle_ts", "internal/oracle/testdata/regexp_served_casts/field_call_angle.ts", "true\n", ":2:33: stage 0 can't lower an opaque host or collection value viewed as unknown or object (dynamic property metadata) yet"},
}

func TestRegExpErasedCastStops(t *testing.T) {
	t.Parallel()
	for _, row := range regexpErasedCastRows {
		t.Run(row.name, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, row.path))
			if err != nil {
				t.Fatal(err)
			}
			truth := execute(t, "node", "--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle/node.mjs"), path)
			if difference := disagreement(run{stdout: []byte(row.stdout)}, truth); difference != "" {
				t.Fatalf("Node: %s: %+v", difference, truth)
			}
			checked, err := load.Load([]string{path})
			if err == nil {
				_, err = lower.Lower(context.Background(), checked)
			}
			if err == nil || err.Error() != path+row.diagnostic {
				t.Fatalf("want exact stop %q, got %v", row.diagnostic, err)
			}
		})
	}
}
