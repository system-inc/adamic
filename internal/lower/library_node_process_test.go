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

func TestNodeProcessDeferredSchedulingStaysLoud(t *testing.T) {
	t.Parallel()
	for _, source := range []string{"process.nextTick(()=>{});", "setTimeout(()=>{},0);"} {
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

// stdoutHandleView is the view tsc's sys.ts casts stdout's handle to (sys.ts:1609).
const stdoutHandleView = "import 'node:process'; const handle = (process.stdout as any)._handle as { setBlocking?: (value: boolean) => void; }; "

// setBlocking read off stdout's handle as a value stops by name, wherever the value goes (#5abf5v8):
// Node's needs the handle as its receiver. The origin is followed through casts and const bindings,
// never read off the view type. A .a program refuses the non-null !, so that road is a .ts row.
func TestNodeProcessHandleSetBlockingValueStops(t *testing.T) {
	t.Parallel()
	for _, row := range []struct{ name, source, extension string }{
		{"bound", "const set = handle.setBlocking; if (set) { set(true); }", ".a"},
		{"argument", "function run(set?: (value: boolean) => void): void { if (set) { set(true); } } run(handle.setBlocking);", ".a"},
		{"returned", "function take(): ((value: boolean) => void) | undefined { return handle.setBlocking; } const set = take(); if (set) { set(true); }", ".a"},
		{"optional chain", "const set = handle?.setBlocking; if (set) { set(true); }", ".a"},
		{"destructured", "const { setBlocking } = handle; if (setBlocking) { setBlocking(true); }", ".a"},
		{"element", "const set = handle['setBlocking']; if (set) { set(true); }", ".a"},
		{"satisfies and as", "const view = (handle satisfies { setBlocking?: (value: boolean) => void; }) as { setBlocking?: (value: boolean) => void; }; const set = view.setBlocking; if (set) { set(true); }", ".a"},
		{"non-null", "const view = handle!; const set = view.setBlocking; if (set) { set(true); }", ".ts"},
	} {
		t.Run(row.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSourceAs(t, stdoutHandleView+row.source, row.extension)
			var missing *NotYet
			if !errors.As(err, &missing) || missing.What != stdoutHandleSetBlockingValue {
				t.Fatalf("want the setBlocking value stop, got %v", err)
			}
		})
	}
}

// A call's callee and a truthiness test read setBlocking without detaching it, so they still lower.
func TestNodeProcessHandleSetBlockingCallAndTestLower(t *testing.T) {
	t.Parallel()
	for name, source := range map[string]string{
		"callee":      "import 'node:process'; const handle = (process.stdout as any)._handle as { setBlocking: (value: boolean) => void; }; handle.setBlocking(true); console.log('after');",
		"if":          stdoutHandleView + "if (handle.setBlocking) { console.log('present'); }",
		"and":         stdoutHandleView + "if (handle && handle.setBlocking) { console.log('present'); }",
		"not":         stdoutHandleView + "if (!handle.setBlocking) { console.log('absent'); }",
		"tsc's guard": stdoutHandleView + "if (handle && handle.setBlocking) { handle.setBlocking(true); }",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if _, err := lowerSource(t, source); err != nil {
				t.Fatal(err)
			}
		})
	}
}
