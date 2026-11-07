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

func TestNodeProcessPerformanceMixedClosures_010(test *testing.T) {
	test.Parallel()
	nodeProcessPerformanceMixedClosures(test, 0)
}
func TestNodeProcessPerformanceMixedClosures_011(test *testing.T) {
	test.Parallel()
	nodeProcessPerformanceMixedClosures(test, 1)
}
func TestNodeProcessPerformanceMixedClosures_012(test *testing.T) {
	test.Parallel()
	nodeProcessPerformanceMixedClosures(test, 2)
}

func nodeProcessPerformanceMixedClosures(test *testing.T, shard int) {
	source := `import {performance} from 'node:perf_hooks';
interface Clock { now():number; mark(name:string):void; measure(name:string,start?:string,end?:string):void; clearMarks(name?:string):void; }
class EarlyCounter { next():number {return 1;} }
function early(value:{next():number}):number {return value.next();}
console.log(String(early(new EarlyCounter())));
const p:Clock=performance;
const counter={value:0, next():number {this.value++; return this.value;}};
function advance(value:{next():number}):number {return value.next();}
console.log(String(advance(counter)));
p.mark('x'); p.measure('x'); p.clearMarks();
console.log(String(p.now()>=0));
`
	path := filepath.Join(test.TempDir(), "main.a")
	if writeError := os.WriteFile(path, []byte(source), 0600); writeError != nil {
		test.Fatal(writeError)
	}
	program, loweredError := lowered(test, path)
	if loweredError != nil {
		test.Fatal(loweredError)
	}
	truth := onNodeWith(test, inputRun{}, path)
	nodeProcessRuntimeAgreement(test, program, truth, shard)
}

func TestNodeProcessPerformanceOptionsMutant_000(test *testing.T) {
	test.Parallel()
	nodeProcessPerformanceOptionsMutant(test, 0)
}
func TestNodeProcessPerformanceOptionsMutant_001(test *testing.T) {
	test.Parallel()
	nodeProcessPerformanceOptionsMutant(test, 1)
}

func nodeProcessPerformanceOptionsMutant(test *testing.T, shard int) {
	source := `import {performance} from 'node:perf_hooks'; const p=performance; const m=p.measure('x',{start:0,end:0}); console.log(String(m.duration===0));`
	path := filepath.Join(test.TempDir(), "main.a")
	if writeError := os.WriteFile(path, []byte(source), 0600); writeError != nil {
		test.Fatal(writeError)
	}
	_, loweredError := lowered(test, path)
	var missing *lower.NotYet
	if !errors.As(loweredError, &missing) {
		test.Fatalf("want explicit options ABI refusal, got %v", loweredError)
	}
	truth := onNodeWith(test, inputRun{}, path)
	mutantPath := filepath.Join(test.TempDir(), "main.a")
	if writeError := os.WriteFile(
		mutantPath,
		[]byte(strings.Replace(source, ",{start:0,end:0}", "", 1)),
		0600,
	); writeError != nil {
		test.Fatal(writeError)
	}
	program, loweredError := lowered(test, mutantPath)
	if loweredError != nil {
		test.Fatal(loweredError)
	}
	var got run
	var binary string
	if shard == 0 {
		got, binary = nativelyUncached(test, program)
	} else {
		got = releasedUncached(test, program)
	}
	if truth.exitCode != 0 ||
		len(truth.stderr) != 0 ||
		got.exitCode != 0 ||
		len(got.stderr) != 0 ||
		disagreement(truth, got) != "stdout differs" {
		test.Fatalf("option-erasure mutant failed outside Node comparison: Node %+v mutant %+v", truth, got)
	}
	if shard == 0 {
		if report := leaksUncached(test, program, binary); report != "" {
			test.Fatal(report)
		}
	}
	test.Log("discarded performance options ran cleanly and were caught only by Node stdout")
}

// This is the old structural-store IR: the object accepts a new method, but a
// subsequent direct host call bypasses the slot. Only Node sees the override.
func TestNodeProcessPerformanceMutationMutant_000(test *testing.T) {
	test.Parallel()
	nodeProcessPerformanceMutationMutant(test, 0)
}
func TestNodeProcessPerformanceMutationMutant_001(test *testing.T) {
	test.Parallel()
	nodeProcessPerformanceMutationMutant(test, 1)
}

func nodeProcessPerformanceMutationMutant(test *testing.T, shard int) {
	source := `import {performance} from 'node:perf_hooks'; const p:{now:()=>number}=performance; p.now=()=>-1; console.log(String(performance.now()===-1));`
	path := filepath.Join(test.TempDir(), "main.a")
	if writeError := os.WriteFile(path, []byte(source), 0600); writeError != nil {
		test.Fatal(writeError)
	}
	_, loweredError := lowered(test, path)
	var missing *lower.NotYet
	if !errors.As(loweredError, &missing) {
		test.Fatalf("want explicit host slot refusal, got %v", loweredError)
	}
	truth := onNodeWith(test, inputRun{}, path)
	program := &ir.Program{Source: "performance-mutation-mutant.a", Locals: []ir.Local{
		{Name: "host", Type: ir.Object, Function: -1},
	}, Functions: []ir.Function{{
		Name:    "replacement",
		Closure: true,
		Returns: ir.Number,
		Body:    []ir.Statement{ir.Return{Value: ir.NumberConstant{Value: -1}}},
	}}}
	program.Main = []ir.Statement{
		ir.Declare{Local: 0, Value: ir.ProcessCall{Operation: "performance", Of: ir.Object}},
		ir.SetProperty{Object: ir.Read{Local: 0, Of: ir.Object}, Name: "now", Value: ir.MakeClosure{Function: 0}},
		ir.WriteLine{Stream: ir.Stdout, Value: ir.BooleanToString{Value: ir.Binary{
			Operator: ir.Equal,
			Left:     ir.ProcessCall{Operation: "now", Of: ir.Number},
			Right:    ir.NumberConstant{Value: -1},
		}}},
	}
	var got run
	var binary string
	if shard == 0 {
		got, binary = nativelyUncached(test, program)
	} else {
		got = releasedUncached(test, program)
	}
	if truth.exitCode != 0 ||
		len(truth.stderr) != 0 ||
		got.exitCode != 0 ||
		len(got.stderr) != 0 ||
		disagreement(truth, got) != "stdout differs" {
		test.Fatalf("static-dispatch mutant failed outside Node comparison: Node %+v mutant %+v", truth, got)
	}
	if shard == 0 {
		if report := leaksUncached(test, program, binary); report != "" {
			test.Fatal(report)
		}
	}
	test.Log("ignored performance override ran cleanly and was caught only by Node stdout")
}
