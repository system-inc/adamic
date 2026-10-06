package oracle

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
)

// A structural type can hide fields, even when all of its visible fields are scalar. These
// programs must be refused until stringify can inspect complete runtime value metadata.
func TestJSONStringifyRefusals(t *testing.T) {
	t.Parallel()
	cases := []struct{ name, source, reason string }{
		{"parse", `JSON.parse('{"n":1}');`, "result's type can't be proven from the text"},
		{"hidden_fields", `const full = { n: 1, extra: 'kept by Node' }; const view: { readonly n: number } = full; JSON.stringify(view);`, "structural types can hide fields"},
		{"array_of_objects", `const items = [{ n: 1 }]; JSON.stringify(items);`, "structural types can hide fields"},
		{"toJSON", `JSON.stringify({ toJSON: () => 7 });`, "toJSON semantics"},
		{"replacer_function", `JSON.stringify(7, (key: string, value: number) => value);`, "callback must have a proven type"},
		{"prototype", `JSON.stringify({ __proto__: null });`, "__proto__ semantics"},
		{"duplicate", `JSON.stringify({ x: 1, x: 2 });`, "JSON.stringify duplicate literal keys"},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "probe.ts")
			if err := os.WriteFile(path, []byte(test.source), 0600); err != nil {
				t.Fatal(err)
			}
			program, err := load.Load([]string{path})
			if err != nil {
				// The checker itself refuses duplicate keys before lowering.
				if test.name == "duplicate" && strings.Contains(err.Error(), "TS1117") {
					return
				}
				t.Fatal(err)
			}
			_, err = lower.Lower(context.Background(), program)
			if err == nil || !strings.Contains(err.Error(), test.reason) {
				t.Fatalf("want refusal containing %q, got %v", test.reason, err)
			}
			if test.name == "parse" && !strings.Contains(err.Error(), "checked parse against a declared type is a later design") {
				t.Fatalf("parse refusal omitted its later design: %v", err)
			}
		})
	}
}

func TestJSONStringifyResultMayBeUndefined(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "probe.ts")
	// Even a serializable argument is not used as a license to lie in the public declaration.
	if err := os.WriteFile(path, []byte(`const result: string = JSON.stringify(undefined); console.log(result);`), 0600); err != nil {
		t.Fatal(err)
	}
	_, err := load.Load([]string{path})
	if err == nil || !strings.Contains(err.Error(), "undefined") {
		t.Fatalf("want the checker to require handling stringify's undefined result, got %v", err)
	}
}

// This mutant compiles, exits zero, and leaks nothing. Only the independent Node comparison can
// notice that the descriptor enumerates integer-index keys in the wrong order.
func TestJSONStringifyOracleCatchesKeyOrder(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "mutant.ts")
	source := `console.log(JSON.stringify({ '2': 2, '1': 1, z: 'built'.repeat(2), a: 3 }) ?? 'missing');`
	if err := os.WriteFile(path, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	call := program.Main[0].(ir.WriteLine).Value.(ir.Coalesce).Value.(ir.JSONStringify)
	call.Schema.Fields[0], call.Schema.Fields[1] = call.Schema.Fields[1], call.Schema.Fields[0]
	native, binary := natively(t, program)
	if native.exitCode != 0 || len(native.stderr) != 0 {
		t.Fatalf("mutant must finish cleanly: exit %d stderr %q", native.exitCode, native.stderr)
	}
	if report := leaks(t, program, binary); report != "" {
		t.Fatalf("mutant must not be caught by the leak check: %s", report)
	}
	if difference := disagreement(onNode(t, path), native); difference != "stdout differs" {
		t.Fatalf("want Node alone to catch key order as stdout differs, got %q", difference)
	}
	t.Log("key-order mutant compiled, exited 0, leaked nothing, and Node caught stdout differs")
}
