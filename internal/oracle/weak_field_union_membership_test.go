package oracle

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

func TestWeakFieldUnionMembership(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"members", "mismatch", "decoder"} {
		t.Run(name, func(t *testing.T) {
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/weak_field_union_membership", name+".ts"))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			if weakFieldReads(program) != 1 {
				t.Fatal("expected one weak_field_read")
			}
			fixed := onNode(t, path)
			if fixed.exitCode != 0 || len(fixed.stderr) != 0 {
				t.Fatalf("Node fixed receiver: %s", weakReadShown(fixed))
			}
			assertWeakUnionBackends(t, program, fixed)
			if name == "decoder" {
				// The decoder's receiver can't be put below the relation: past its first read the signature is
				// narrowed to string and read as a fixed object's field, which a record there stops loudly as a
				// compiler bug, so a record substituted in IR proves nothing about this head. Its fixed receiver
				// holds the three branches to Node, and the read the head lowers is the first.
				return
			}

			// The source relation from a record to a weak type stays refused. As in
			// TestWeakReadRecordPath, pass the real record below that relation in IR.
			// Node can run the same indexed writes and calls with the source argument changed.
			var receiver ir.Expression
			for local, binding := range program.Locals {
				if binding.Name == "table" && binding.Type == ir.Record {
					receiver = ir.Narrow{Value: ir.Box{Value: ir.Read{Local: local, Of: ir.Record}}, To: ir.Object}
					break
				}
			}
			if receiver == nil {
				t.Fatal("missing record local table")
			}
			replaced := 0
			rewrite := func(expression ir.Expression) ir.Expression {
				call, ok := expression.(ir.Call)
				if !ok || program.Functions[call.Function].Name != "show" {
					return expression
				}
				if _, absent := call.Arguments[0].(ir.Undefined); absent {
					return expression
				}
				call.Arguments = []ir.Expression{receiver}
				replaced++
				return call
			}
			mutateStringExpressions(reflect.ValueOf(&program.Main).Elem(), rewrite)
			if replaced != map[string]int{"members": 4, "mismatch": 1}[name] {
				t.Fatalf("replaced %d fixed receivers", replaced)
			}

			source, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			lines := strings.Split(string(source), "\n")
			for index, line := range lines {
				if strings.HasPrefix(line, "show({") {
					lines[index] = "show(table as unknown as Fields);"
				}
			}
			recordPath := filepath.Join(t.TempDir(), name+".ts")
			if err := os.WriteFile(recordPath, []byte(strings.Join(lines, "\n")), 0o644); err != nil {
				t.Fatal(err)
			}
			truth := onNode(t, recordPath)
			t.Logf("Node record receiver: %s", weakReadShown(truth))
			if truth.exitCode != 0 || len(truth.stderr) != 0 {
				t.Fatalf("Node record receiver: %s", weakReadShown(truth))
			}
			want := truth
			if name == "mismatch" {
				if string(truth.stdout) != "true\n" {
					t.Fatalf("Node must print the indexed write: %s", weakReadShown(truth))
				}
				const message = "adamic: panic: field read failed: fields.value is not a string | number | null | undefined; expected string | number | null | undefined, found boolean, which a computed-key or index-signature write may have stored\n"
				want = run{exitCode: 70, stderr: []byte(message)}
			} else if string(truth.stdout) != "entry\n7\nnull\nundefined\nundefined\n" {
				t.Fatalf("Node changed the record values or absence: %s", weakReadShown(truth))
			}
			assertWeakUnionBackends(t, program, want)
		})
	}
}

func assertWeakUnionBackends(t *testing.T, program *ir.Program, want run) {
	t.Helper()
	for backend, got := range map[string]run{
		"JavaScript": onJavaScriptBackend(t, program),
		"release":    released(t, program),
	} {
		t.Logf("%s: %s", backend, weakReadShown(got))
		if difference := disagreement(want, got); difference != "" {
			t.Fatalf("%s: %s: want %s; got %s", backend, difference, weakReadShown(want), weakReadShown(got))
		}
	}
}
