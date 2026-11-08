package worker

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
	"path/filepath"
	"strings"
	"testing"
)

func TestWasmPurityRefusesEnvironmentMutations(test *testing.T) {
	test.Parallel()
	for _, one := range []struct{ operation, source string }{
		{"host_env_set", `process.env.ADAMIC_WORKER_ENV = 'a';`},
		{"host_env_delete", `delete process.env.ADAMIC_WORKER_ENV;`},
	} {
		test.Run(one.operation, func(test *testing.T) {
			test.Parallel()
			entry := filepath.Join(test.TempDir(), "environment.a")
			write(
				test,
				entry,
				`import process from 'node:process'; import type {HttpRequest,HttpResponse} from 'adamic/http';
export function mutate(value: number): number {`+one.source+` return value;}
export function handle(request: HttpRequest): HttpResponse {return {status:200,headers:[],body:'ok'};}`,
			)
			builds := 0
			buildWithError := BuildWith(
				entry,
				filepath.Join(test.TempDir(), "worker"),
				Options{Wasm: []string{"mutate"}, BuildWasm: func(
					string,
					string,
					string,
					[]string,
				) error {
					builds++
					return fmt.Errorf("unexpected Wasm build")
				}},
			)
			if buildWithError == nil ||
				!strings.Contains(buildWithError.Error(), "impure: host operation "+one.operation) ||
				builds != 0 {
				test.Fatalf("want environment refusal before Wasm build, builds=%d: %v", builds, buildWithError)
			}
		})
	}
}

func TestCheckPureRefusesEnvironmentHostIr(test *testing.T) {
	test.Parallel()
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
		test.Run(one.name, func(test *testing.T) {
			test.Parallel()
			program := &ir.Program{
				Functions: []ir.Function{{Name: "mutate", Body: []ir.Statement{ir.Evaluate{Value: one.value}}}},
			}
			if checkPureError := CheckPure(
				program,
				0,
			); checkPureError == nil || !strings.Contains(checkPureError.Error(), "impure: host operation") {
				test.Fatalf("want host effect refusal, got %v", checkPureError)
			}
		})
	}
}
