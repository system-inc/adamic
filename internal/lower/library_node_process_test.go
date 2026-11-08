package lower

import (
	"errors"
	"strings"
	"testing"
)

func TestNodeProcessPerformanceAliasAbi(test *testing.T) {
	test.Parallel()
	for name, source := range map[string]string{
		"direct options": "performance.measure('x',{start:0});",
		"alias options":  "const p=performance; p.measure('x',{start:0,end:0});",
		"mark options":   "const p=performance; p.mark('x',{startTime:0});",
	} {
		test.Run(name, func(test *testing.T) {
			test.Parallel()
			_, lowerSourceError := lowerSource(test, "import {performance} from 'node:perf_hooks'; "+source)
			var missing *NotYet
			if !errors.As(lowerSourceError, &missing) || !strings.Contains(missing.What, "performance.") {
				test.Fatalf("want performance ABI refusal, got %v", lowerSourceError)
			}
		})
	}
	for name, source := range map[string]string{
		"alias strings":  "const p=performance; p.mark('x'); p.measure('x'); p.clearMarks(); p.clearMeasures(undefined); console.log(String(p.now()));",
		"void interface": "const p: {mark(name:string):void; measure(name:string,start?:string,end?:string):void; clearMarks(name?:string):void}=performance; p.mark('x'); p.measure('x'); p.clearMarks();",
	} {
		test.Run(name, func(test *testing.T) {
			test.Parallel()
			if _, lowerSourceError := lowerSource(
				test,
				"import {performance} from 'node:perf_hooks'; "+source,
			); lowerSourceError != nil {
				test.Fatal(lowerSourceError)
			}
		})
	}
}

func TestNodeProcessDeferredSchedulingStaysLoud(test *testing.T) {
	test.Parallel()
	for _, source := range []string{"process.nextTick(()=>{});", "setTimeout(()=>{},0);"} {
		test.Run(source, func(test *testing.T) {
			test.Parallel()
			if _, lowerSourceError := lowerSource(test, "import 'node:process'; "+source); lowerSourceError == nil {
				test.Fatal("unsupported host operation silently compiled")
			}
		})
	}
}

func TestNodeProcessPerformanceStructuralBoundaries(test *testing.T) {
	test.Parallel()
	for _, source := range []string{
		"performance.now=()=>17; console.log(String(performance.now()));",
		"const p:{now:()=>number}=performance; p.now=()=>17; console.log(String(performance.now()));",
		"const p:{mark(name:string,options?:{startTime:number}):void}=performance; p.mark('x',{startTime:0});",
	} {
		test.Run(source, func(test *testing.T) {
			test.Parallel()
			_, lowerSourceError := lowerSource(test, "import {performance} from 'node:perf_hooks'; "+source)
			if lowerSourceError == nil {
				test.Fatal("performance mutation/options escaped ABI proof")
			}
		})
	}
}
