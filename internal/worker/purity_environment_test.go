package worker

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
	"path/filepath"
	"strings"
	"testing"
)

func TestWasmPurityRefusesEnvironmentMutations(t *testing.T) {
	t.Parallel()
	for _, one := range []struct{ operation, source string }{
		{"host_env_set", `process.env.ADAMIC_WORKER_ENV = 'a';`},
		{"host_env_delete", `delete process.env.ADAMIC_WORKER_ENV;`},
	} {
		t.Run(one.operation, func(t *testing.T) {
			t.Parallel()
			entry := filepath.Join(t.TempDir(), "environment.a")
			write(t, entry, `import process from 'node:process'; import type {HttpRequest,HttpResponse} from 'adamic/http';
export function mutate(value: number): number {`+one.source+` return value;}
export function handle(request: HttpRequest): HttpResponse {return {status:200,headers:[],body:'ok'};}`)
			builds := 0
			err := BuildWith(entry, filepath.Join(t.TempDir(), "worker"), Options{Wasm: []string{"mutate"}, BuildWasm: func(string, string, string, []string) error { builds++; return fmt.Errorf("unexpected Wasm build") }})
			if err == nil || !strings.Contains(err.Error(), "impure: host operation "+one.operation) || builds != 0 {
				t.Fatalf("want environment refusal before Wasm build, builds=%d: %v", builds, err)
			}
		})
	}
}

func TestCheckPureRefusesEnvironmentHostIR(t *testing.T) {
	t.Parallel()
	for _, one := range []struct {
		name  string
		value ir.Expression
	}{
		{"file set", ir.NodeFSFile{Operation: "host_env_set", Of: ir.String}},
		{"file delete", ir.NodeFSFile{Operation: "host_env_delete", Of: ir.Boolean}},
		{"host set", ir.NodeHostCall{Module: "node:process", Member: "envSet", Returns: ir.String}},
		{"host delete", ir.NodeHostCall{Module: "node:process", Member: "envDelete", Returns: ir.Boolean}},
		{"process host value", ir.ProcessCall{Operation: "performance", Of: ir.Object}},
	} {
		t.Run(one.name, func(t *testing.T) {
			t.Parallel()
			program := &ir.Program{Functions: []ir.Function{{Name: "mutate", Body: []ir.Statement{ir.Evaluate{Value: one.value}}}}}
			if err := CheckPure(program, 0); err == nil || !strings.Contains(err.Error(), "impure: host operation") {
				t.Fatalf("want host effect refusal, got %v", err)
			}
		})
	}
}
