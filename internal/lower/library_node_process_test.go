package lower

import (
	"errors"
	"strings"
	"testing"
)

func TestNodeProcessPerformanceAliasABI(t *testing.T) {
	t.Parallel()
	for name, source := range map[string]string{
		"direct options": "performance.measure('x',{start:0});",
		"alias options":  "const p=performance; p.measure('x',{start:0,end:0});",
		"mark options":   "const p=performance; p.mark('x',{startTime:0});",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSource(t, "import {performance} from 'node:perf_hooks'; "+source)
			var missing *NotYet
			if !errors.As(err, &missing) || !strings.Contains(missing.What, "performance.") {
				t.Fatalf("want performance ABI refusal, got %v", err)
			}
		})
	}
	for name, source := range map[string]string{
		"alias strings":  "const p=performance; p.mark('x'); p.measure('x'); p.clearMarks(); p.clearMeasures(undefined); console.log(String(p.now()));",
		"void interface": "const p: {mark(name:string):void; measure(name:string,start?:string,end?:string):void; clearMarks(name?:string):void}=performance; p.mark('x'); p.measure('x'); p.clearMarks();",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if _, err := lowerSource(t, "import {performance} from 'node:perf_hooks'; "+source); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestNodeProcessDeferredExitStaysLoud(t *testing.T) {
	t.Parallel()
	for _, source := range []string{"process.exit(0);", "process.exitCode=1;", "process.nextTick(()=>{});", "setTimeout(()=>{},0);"} {
		t.Run(source, func(t *testing.T) {
			t.Parallel()
			if _, err := lowerSource(t, "import 'node:process'; "+source); err == nil {
				t.Fatal("unsupported host operation silently compiled")
			}
		})
	}
}

func TestNodeProcessPerformanceStructuralBoundaries(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"performance.now=()=>17; console.log(String(performance.now()));",
		"const p:{now:()=>number}=performance; p.now=()=>17; console.log(String(performance.now()));",
		"const p:{mark(name:string,options?:{startTime:number}):void}=performance; p.mark('x',{startTime:0});",
	} {
		t.Run(source, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSource(t, "import {performance} from 'node:perf_hooks'; "+source)
			if err == nil {
				t.Fatal("performance mutation/options escaped ABI proof")
			}
		})
	}
}

func TestNodeProcessTTYFallbackABI(t *testing.T) {
	t.Parallel()
	for _, source := range []string{"process.stdout.isTTY??'missing'", "process.stdout.columns??false"} {
		t.Run(source, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSource(t, "import 'node:process'; console.log(String("+source+"));")
			var missing *NotYet
			if !errors.As(err, &missing) {
				t.Fatalf("want named TTY ABI refusal, got %v", err)
			}
		})
	}
	for _, source := range []string{"process.stdout.isTTY??false", "process.stdout.columns??-1"} {
		t.Run(source, func(t *testing.T) {
			t.Parallel()
			if _, err := lowerSource(t, "import 'node:process'; console.log(String("+source+"));"); err != nil {
				t.Fatal(err)
			}
		})
	}
}
