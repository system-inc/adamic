package oracle

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/lower"
)

func TestNodeProcessPerformanceMixedClosures_010(t *testing.T) {
	nodeProcessPerformanceMixedClosures(t, 0)
}
func TestNodeProcessPerformanceMixedClosures_011(t *testing.T) {
	nodeProcessPerformanceMixedClosures(t, 1)
}
func TestNodeProcessPerformanceMixedClosures_012(t *testing.T) {
	nodeProcessPerformanceMixedClosures(t, 2)
}

func nodeProcessPerformanceMixedClosures(t *testing.T, shard int) {
	t.Parallel()
	source := `import {performance} from 'node:perf_hooks';
interface Clock { now():number; mark(name:string):void; measure(name:string,start?:string,end?:string):void; clearMarks(name?:string):void; }
const p:Clock=performance;
const counter={value:0, next():number {this.value++; return this.value;}};
function advance(value:{next():number}):number {return value.next();}
console.log(String(advance(counter)));
p.mark('x'); p.measure('x'); p.clearMarks();
console.log(String(p.now()>=0));
`
	path := filepath.Join(t.TempDir(), "main.a")
	if err := os.WriteFile(path, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	truth := onNodeWith(t, inputRun{}, path)
	nodeProcessRuntimeAgreement(t, program, truth, shard)
}

func TestNodeProcessPerformanceOptionsMutant_000(t *testing.T) {
	nodeProcessPerformanceOptionsMutant(t, 0)
}
func TestNodeProcessPerformanceOptionsMutant_001(t *testing.T) {
	nodeProcessPerformanceOptionsMutant(t, 1)
}

func nodeProcessPerformanceOptionsMutant(t *testing.T, shard int) {
	t.Parallel()
	source := `import {performance} from 'node:perf_hooks'; const p=performance; const m=p.measure('x',{start:0,end:0}); console.log(String(m.duration===0));`
	path := filepath.Join(t.TempDir(), "main.a")
	if err := os.WriteFile(path, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	_, err := lowered(t, path)
	var missing *lower.NotYet
	if !errors.As(err, &missing) {
		t.Fatalf("want explicit options ABI refusal, got %v", err)
	}
	truth := onNodeWith(t, inputRun{}, path)
	mutantPath := filepath.Join(t.TempDir(), "main.a")
	if err := os.WriteFile(mutantPath, []byte(strings.Replace(source, ",{start:0,end:0}", "", 1)), 0600); err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, mutantPath)
	if err != nil {
		t.Fatal(err)
	}
	var got run
	var binary string
	if shard == 0 {
		got, binary = nativelyUncached(t, program)
	} else {
		got = releasedUncached(t, program)
	}
	if truth.exitCode != 0 || len(truth.stderr) != 0 || got.exitCode != 0 || len(got.stderr) != 0 || disagreement(truth, got) != "stdout differs" {
		t.Fatalf("option-erasure mutant failed outside Node comparison: Node %+v mutant %+v", truth, got)
	}
	if shard == 0 {
		if report := leaksUncached(t, program, binary); report != "" {
			t.Fatal(report)
		}
	}
	t.Log("discarded performance options ran cleanly and were caught only by Node stdout")
}

// This is the old structural-store IR: the object accepts a new method, but a
// subsequent direct host call bypasses the slot. Only Node sees the override.
func TestNodeProcessPerformanceMutationMutant_000(t *testing.T) {
	nodeProcessPerformanceMutationMutant(t, 0)
}
func TestNodeProcessPerformanceMutationMutant_001(t *testing.T) {
	nodeProcessPerformanceMutationMutant(t, 1)
}

func nodeProcessPerformanceMutationMutant(t *testing.T, shard int) {
	t.Parallel()
	source := `import {performance} from 'node:perf_hooks'; const p:{now:()=>number}=performance; p.now=()=>-1; console.log(String(performance.now()===-1));`
	path := filepath.Join(t.TempDir(), "main.a")
	if err := os.WriteFile(path, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	_, err := lowered(t, path)
	var missing *lower.NotYet
	if !errors.As(err, &missing) {
		t.Fatalf("want explicit host slot refusal, got %v", err)
	}
	truth := onNodeWith(t, inputRun{}, path)
	program := &ir.Program{Source: "performance-mutation-mutant.a", Locals: []ir.Local{{Name: "host", Type: ir.Object, Function: -1}}, Functions: []ir.Function{{Name: "replacement", Closure: true, Returns: ir.Number, Body: []ir.Statement{ir.Return{Value: ir.NumberConstant{Value: -1}}}}}}
	program.Main = []ir.Statement{
		ir.Declare{Local: 0, Value: ir.ProcessCall{Operation: "performance", Of: ir.Object}},
		ir.SetProperty{Object: ir.Read{Local: 0, Of: ir.Object}, Name: "now", Value: ir.MakeClosure{Function: 0}},
		ir.WriteLine{Stream: ir.Stdout, Value: ir.BooleanToString{Value: ir.Binary{Operator: ir.Equal, Left: ir.ProcessCall{Operation: "now", Of: ir.Number}, Right: ir.NumberConstant{Value: -1}}}},
	}
	var got run
	var binary string
	if shard == 0 {
		got, binary = nativelyUncached(t, program)
	} else {
		got = releasedUncached(t, program)
	}
	if truth.exitCode != 0 || len(truth.stderr) != 0 || got.exitCode != 0 || len(got.stderr) != 0 || disagreement(truth, got) != "stdout differs" {
		t.Fatalf("static-dispatch mutant failed outside Node comparison: Node %+v mutant %+v", truth, got)
	}
	if shard == 0 {
		if report := leaksUncached(t, program, binary); report != "" {
			t.Fatal(report)
		}
	}
	t.Log("ignored performance override ran cleanly and was caught only by Node stdout")
}
