package oracle

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
)

var regexpServedValueRows = []struct{ name, path, stdout, diagnostic string }{
	{"logical_null", "internal/oracle/testdata/regexp_served_values/logical_null.a", "null\n", ":3:47: stage 0 can't lower String conversion of an object, array, map or function (ToPrimitive is not lowered) yet"},
	{"logical_undefined", "internal/oracle/testdata/regexp_served_values/logical_undefined.a", "undefined\n", ":3:47: stage 0 can't lower String conversion of an object, array, map or function (ToPrimitive is not lowered) yet"},
	{"logical_null_narrowed", "internal/oracle/testdata/regexp_served_values/logical_null_narrowed.a", "null\n", ":3:43: stage 0 can't lower a presence test of optional member test that lowering didn't leave as one field read yet"},
	{"logical_undefined_narrowed", "internal/oracle/testdata/regexp_served_values/logical_undefined_narrowed.a", "undefined\n", ":3:43: stage 0 can't lower a presence test of optional member test that lowering didn't leave as one field read yet"},
	{"host_handle", "internal/oracle/testdata/regexp_served_values/host_handle.a", "false\n", ":4:14: stage 0 can't lower process.stdout._handle seen as M, which names test where the runtime's stdout handle holds only setBlocking yet"},
	{"host_stats_reach", "internal/oracle/testdata/regexp_served_values/host_stats_reach.a", "false\n", ":3:20: Adamic 0.1 refuses optional property size in M absent from structural source RegExp, which can hide fields; declare size on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"},
	{"no_relation_presence", "internal/oracle/testdata/regexp_served_values/no_relation_presence.a", "false\ntrue\nhas test\n1\ntrue\nfalse\nundefined\n", ""},
	{"mapped_trigger", "internal/oracle/testdata/regexp_served_values/mapped_trigger.a", "true\n", ":4:11: stage 0 can't lower test read as a value, in a program where a RegExp reaches a view naming test, which Adamic can't give a receiver to yet"},
	{"other_key", "internal/oracle/testdata/regexp_served_values/other_key.a", "a\n", ""},
	{"record_dot", "internal/oracle/testdata/regexp_served_values/record_dot.a", "function\n", ":4:11: stage 0 can't lower test read as a value, in a program where a RegExp reaches a view naming test, which Adamic can't give a receiver to yet"},
	{"record_key", "internal/oracle/testdata/regexp_served_values/record_key.a", "function\n", ":4:11: stage 0 can't lower test read as a value, in a program where a RegExp reaches a view naming test, which Adamic can't give a receiver to yet"},
	{"alias", "internal/oracle/testdata/regexp_served_values/alias.a", "function\n", ":5:11: stage 0 can't lower test read as a value, in a program where a RegExp reaches a view naming test, which Adamic can't give a receiver to yet"},
	{"also_call", "internal/oracle/testdata/regexp_served_values/also_call.a", "true\ntrue\n", ""},
	{"also_presence", "internal/oracle/testdata/regexp_served_values/also_presence.a", "false\nhas test\ntrue\nno test\n", ""},
	{"also_value", "internal/oracle/testdata/regexp_served_values/also_value.a", "function\n", ":5:11: stage 0 can't lower test read as a value, in a program where a RegExp reaches a view naming test, which Adamic can't give a receiver to yet"},
	{"array", "internal/oracle/testdata/regexp_served_values/array.a", "function\n", ":4:34: stage 0 can't lower test read as a value, in a program where a RegExp reaches a view naming test, which Adamic can't give a receiver to yet"},
	{"assign", "internal/oracle/testdata/regexp_served_values/assign.a", "missing\n", ":4:1: stage 0 can't lower test read as a value, in a program where a RegExp reaches a view naming test, which Adamic can't give a receiver to yet"},
	{"before", "internal/oracle/testdata/regexp_served_values/before.a", "function\ntrue\n", ":2:43: stage 0 can't lower test read as a value, in a program where a RegExp reaches a view naming test, which Adamic can't give a receiver to yet"},
	{"class_field", "internal/oracle/testdata/regexp_served_values/class_field.a", "function\n", ":3:20: stage 0 can't lower test read as a value, in a program where a RegExp reaches a view naming test, which Adamic can't give a receiver to yet"},
	{"class_presence", "internal/oracle/testdata/regexp_served_values/class_presence.a", "true\ntruthy\ntrue\ntruthy\nfalse\n", ":3:51: stage 0 can't lower test read as a value, in a program where a RegExp reaches a view naming test, which Adamic can't give a receiver to yet"},
	{"destructuring", "internal/oracle/testdata/regexp_served_values/destructuring.a", "function\n", ":3:7: stage 0 can't lower test read as a value, in a program where a RegExp reaches a view naming test, which Adamic can't give a receiver to yet"},
	{"element", "internal/oracle/testdata/regexp_served_values/element.a", "function\n", ":3:11: stage 0 can't lower test read as a value, in a program where a RegExp reaches a view naming test, which Adamic can't give a receiver to yet"},
	{"entries", "internal/oracle/testdata/regexp_served_values/entries.a", "0\n", ":3:20: stage 0 can't lower test read as a value, in a program where a RegExp reaches a view naming test, which Adamic can't give a receiver to yet"},
	{"generic", "internal/oracle/testdata/regexp_served_values/generic.a", "function\n", ":2:56: stage 0 can't lower test read as a value, in a program where a RegExp reaches a view naming test, which Adamic can't give a receiver to yet"},
	{"logical_assignment", "internal/oracle/testdata/regexp_served_values/logical_assignment.a", "function\n", ":3:12: stage 0 can't lower test read as a value, in a program where a RegExp reaches a view naming test, which Adamic can't give a receiver to yet"},
	{"object_literal", "internal/oracle/testdata/regexp_served_values/object_literal.a", "function\n", ":3:21: stage 0 can't lower test read as a value, in a program where a RegExp reaches a view naming test, which Adamic can't give a receiver to yet"},
	{"optional_view", "internal/oracle/testdata/regexp_served_values/optional_view.a", "function\n", ":4:34: stage 0 can't lower test read as a value, in a program where a RegExp reaches a view naming test, which Adamic can't give a receiver to yet"},
	{"over_stop", "internal/oracle/testdata/regexp_served_values/over_stop.a", "true\n", ":4:11: stage 0 can't lower test read as a value, in a program where a RegExp reaches a view naming test, which Adamic can't give a receiver to yet"},
	{"plain", "internal/oracle/testdata/regexp_served_values/plain.a", "true\n", ""},
	{"presence", "internal/oracle/testdata/regexp_served_values/presence.a", "false\ntrue\ntruthy\nfalse\ntrue\ntrue\nfalse\nfalsy\ntrue\nundefined\n", ""},
	{"presence_narrowed", "internal/oracle/testdata/regexp_served_values/presence_narrowed.a", "true\n", ""},
	{"receiver_any", "internal/oracle/testdata/regexp_served_values/receiver_any.a", "true\n", ":3:15: Adamic 0.1 refuses a value of type any: any has no proven type; type it unknown and narrow it where it's read, or give it its real type"},
	{"receiver_intersection", "internal/oracle/testdata/regexp_served_values/receiver_intersection.a", "function\n", ":4:7: stage 0 can't lower a value of type Both yet"},
	{"receiver_parameter", "internal/oracle/testdata/regexp_served_values/receiver_parameter.a", "function\n", ":3:56: stage 0 can't lower test read as a value, in a program where a RegExp reaches a view naming test, which Adamic can't give a receiver to yet"},
	{"receiver_regexp", "internal/oracle/testdata/regexp_served_values/receiver_regexp.a", "true\n", ""},
	{"receiver_union", "internal/oracle/testdata/regexp_served_values/receiver_union.a", "function\n", ":3:52: stage 0 can't lower test read as a value, in a program where a RegExp reaches a view naming test, which Adamic can't give a receiver to yet"},
	{"receiver_unknown", "internal/oracle/testdata/regexp_served_values/receiver_unknown.a", "true\n", ":3:54: error TS18046: 'v' is of type 'unknown'."},
	{"record", "internal/oracle/testdata/regexp_served_values/record.a", "true\n", ":4:39: stage 0 can't lower test read as a value, in a program where a RegExp reaches a view naming test, which Adamic can't give a receiver to yet"},
	{"reflect", "internal/oracle/testdata/regexp_served_values/reflect.a", "function\n", ":3:11: stage 0 can't lower test read as a value, in a program where a RegExp reaches a view naming test, which Adamic can't give a receiver to yet"},
	{"spread", "internal/oracle/testdata/regexp_served_values/spread.a", "missing\n", ":3:13: stage 0 can't lower test read as a value, in a program where a RegExp reaches a view naming test, which Adamic can't give a receiver to yet"},
	{"union", "internal/oracle/testdata/regexp_served_values/union.a", "function\n", ":3:29: stage 0 can't lower test read as a value, in a program where a RegExp reaches a view naming test, which Adamic can't give a receiver to yet"},
	{"values", "internal/oracle/testdata/regexp_served_values/values.a", "0\n", ":3:20: stage 0 can't lower test read as a value, in a program where a RegExp reaches a view naming test, which Adamic can't give a receiver to yet"},
	{"variable", "internal/oracle/testdata/regexp_served_values/variable.a", "function\n", ":3:11: stage 0 can't lower test read as a value, in a program where a RegExp reaches a view naming test, which Adamic can't give a receiver to yet"},
	{"whole_assign", "internal/oracle/testdata/regexp_served_values/whole_assign.a", "true\n", ":5:1: stage 0 can't lower test read as a value, in a program where a RegExp reaches a view naming test, which Adamic can't give a receiver to yet"},
	{"whole_destructuring", "internal/oracle/testdata/regexp_served_values/whole_destructuring.a", "true\n", ":4:7: stage 0 can't lower test read as a value, in a program where a RegExp reaches a view naming test, which Adamic can't give a receiver to yet"},
	{"whole_destructuring_other", "internal/oracle/testdata/regexp_served_values/whole_destructuring_other.a", "1\n", ":4:7: stage 0 can't lower test read as a value, in a program where a RegExp reaches a view naming test, which Adamic can't give a receiver to yet"},
	{"whole_entries", "internal/oracle/testdata/regexp_served_values/whole_entries.a", "1\n", ":4:20: stage 0 can't lower test read as a value, in a program where a RegExp reaches a view naming test, which Adamic can't give a receiver to yet"},
	{"whole_reflect", "internal/oracle/testdata/regexp_served_values/whole_reflect.a", "true\n", ":4:20: stage 0 can't lower test read as a value, in a program where a RegExp reaches a view naming test, which Adamic can't give a receiver to yet"},
	{"whole_spread", "internal/oracle/testdata/regexp_served_values/whole_spread.a", "true\n", ":4:16: stage 0 can't lower test read as a value, in a program where a RegExp reaches a view naming test, which Adamic can't give a receiver to yet"},
	{"whole_values", "internal/oracle/testdata/regexp_served_values/whole_values.a", "1\n", ":4:20: stage 0 can't lower test read as a value, in a program where a RegExp reaches a view naming test, which Adamic can't give a receiver to yet"},
	{"detached", "internal/oracle/testdata/regexp_structural/detached.a", "TypeError: Method RegExp.prototype.test called on incompatible receiver undefined\n", ":3:11: stage 0 can't lower test read as a value, in a program where a RegExp reaches a view naming test, which Adamic can't give a receiver to yet"},
}

func init() {
	for _, row := range regexpServedValueRows {
		if row.diagnostic == "" {
			fixtures = append(fixtures, struct {
				path            string
				lowers, checked bool
			}{row.path, true, false})
		}
	}
}

// Emptying the served table reds TestRegExpStructuralMethods/method and /optional.
// Removing the value stop reds detached and record here. Judging by a recorded view type
// again reds also_value: the second view lowers and native prints missing. Whole readers
// use the same program-wide rule, including an unrelated object's own test.
func TestRegExpServedValueRoads(t *testing.T) {
	t.Parallel()
	for _, row := range regexpServedValueRows {
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
			if row.diagnostic != "" {
				if err == nil {
					_, err = lower.Lower(context.Background(), checked)
				}
				if err == nil || err.Error() != path+row.diagnostic {
					t.Fatalf("want exact stop %q, got %v", row.diagnostic, err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			program, err := lower.Lower(context.Background(), checked)
			if err != nil {
				t.Fatal(err)
			}
			native, binary := nativelyUncached(t, program)
			for name, got := range map[string]run{"sanitized": native, "release": released(t, program), "javascript": onJavaScriptBackend(t, program)} {
				if difference := disagreement(truth, got); difference != "" {
					t.Errorf("%s: %s: exit %d stdout %q stderr %q", name, difference, got.exitCode, got.stdout, got.stderr)
				}
			}
			if report := leaksUncached(t, program, binary); report != "" {
				t.Error(report)
			}
		})
	}
}
