package worker

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
)

func TestWorkerEnvironmentShapeStops(t *testing.T) {
	t.Parallel()
	for _, pin := range []struct{ name, kind, form string }{
		{"e01", "notyet", "Object.keys"}, {"e02", "notyet", "Object.keys"}, {"e10", "notyet", "Marker"},
		{"in", "notyet", "in"}, {"keys", "notyet", "Object.keys"},
		{"values", "notyet", "Object.values"}, {"entries", "notyet", "Object.entries"},
		{"for_in", "notyet", "for...in"}, {"spread", "notyet", "spread"},
		{"json", "notyet", "JSON.stringify"}, {"has_own", "notyet", "hasOwnProperty"},
		{"reflect", "notyet", "Reflect.ownKeys"}, {"object", "notyet", "object"},
		{"array_unknown", "notyet", "unknown"}, {"unknown", "notyet", "unknown"}, {"record", "notyet", "unknown"},
		{"structural", "notyet", "ZONE"}, {"union", "notyet", "Environment | undefined"},
		{"alias_keys", "notyet", "Object.keys"}, {"let", "refused", "slot"},
		{"array", "refused", "slot"}, {"class_field", "refused", "slot"},
		{"parameter_write", "refused", "destructured parameter field"},
		{"optional_parameter", "refused", "slot"}, {"module_binding", "refused", "slot"},
		{"array_temp", "refused", "array element slot"}, {"object_temp", "refused", "object field slot"}, {"generic_new", "notyet", "generic argument"}, {"generic_rest", "notyet", "generic argument"},
		{"return", "refused", "slot"}, {"promise_return", "refused", "slot"},
		{"rest_destructure", "notyet", "spread"}, {"generic_argument", "notyet", "generic argument"}, {"this_parameter", "refused", "slot"},
	} {
		t.Run(pin.name, func(t *testing.T) {
			t.Parallel()
			program, err := load.Load([]string{filepath.Join("testdata/environment_shape", pin.name+".a")})
			if err != nil {
				t.Fatal(err)
			}
			types, release := program.Checker(context.Background(), program.Files()[0])
			_, environment, err := checkedWorkerEnvironment(types, program)
			if err != nil {
				release()
				t.Fatal(err)
			}
			err = lower.CheckWorkerEnvironment(program, types, environment)
			release()
			var notyet *lower.NotYet
			var refused *lower.Refused
			valid := (pin.kind == "notyet" && errors.As(err, &notyet)) || (pin.kind == "refused" && errors.As(err, &refused))
			if !valid || !strings.Contains(err.Error(), pin.form) {
				// The class-field mutant must also pass ordinary lowering, not merely
				// another preflight. Print that outcome at this pin's catch line.
				if pin.name == "class_field" && err == nil {
					_, ordinary := lower.Lower(context.Background(), program)
					t.Logf("class-field road ordinary lowering: %v", ordinary)
				}
				t.Fatalf("environment boundary %s want %s naming %q, got %v", pin.name, pin.kind, pin.form, err)
			}
			if notyet != nil && !strings.Contains(notyet.What, "waits on a bridge that passes Workers' own env") {
				t.Fatalf("environment lifting condition missing: %v", err)
			}
			t.Log(err)
		})
	}
}

func TestWorkerEnvironmentFieldReadRoads(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"const_destructure", "parameter_destructure", "optional_read", "exact_alias", "exact_argument", "unrelated_interface", "async_field_read"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			out := filepath.Join(t.TempDir(), "worker")
			if err := Build(filepath.Join("testdata/environment_shape", name+".a"), out); err != nil {
				t.Fatalf("environment field read build: %v", err)
			}
			script := `import assert from 'node:assert/strict';const {default:w}=await import(process.argv[1]);const r=await w.fetch(new Request('https://test/ok'),{ZONE:'z',GREETING:'hi',EXTRA:'secret'});assert.equal(r.status,200);assert.equal(await r.text(),process.argv[2]);`
			want := "/ok"
			if name == "parameter_destructure" || name == "async_field_read" {
				want = "zhi/ok"
			}
			cmd := exec.CommandContext(t.Context(), "node", "--input-type=module", "-e", script, filepath.Join(out, "worker.mjs"), want)
			if report, err := cmd.CombinedOutput(); err != nil {
				t.Fatalf("environment field read execution: %v\n%s", err, report)
			}
		})
	}
}

func TestWorkerEnvironmentShapeRefusalWritesNothing(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"keys", "class_field", "parameter_write"} {
		out := filepath.Join(t.TempDir(), "worker")
		if err := Build(filepath.Join("testdata/environment_shape", name+".a"), out); err == nil {
			t.Fatalf("environment %s build admitted", name)
		}
		// Build's no-output contract is shared with the signature and bridge checks.
		if _, err := os.Stat(out); !os.IsNotExist(err) {
			t.Fatalf("environment refusal wrote output: %v", err)
		}
	}
}

func TestWorkerEnvironmentMissingTypeIsCompilerBug(t *testing.T) {
	t.Parallel()
	program, err := load.Load([]string{filepath.Join("testdata/environment_shape", "class_field.a")})
	if err != nil {
		t.Fatal(err)
	}
	types, release := program.Checker(context.Background(), program.Files()[0])
	defer release()
	err = lower.CheckWorkerEnvironment(program, types, nil)
	const want = "worker: compiler bug: two-parameter handle has no checked environment type"
	if err == nil || err.Error() != want {
		t.Fatalf("missing checked environment: got %v, want %s", err, want)
	}
}
