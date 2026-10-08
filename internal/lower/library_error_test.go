package lower

import (
	"errors"
	"strings"
	"testing"
)

func TestErrorCaptureBoundaries(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ source, reason string }{
		{`const target={stack:'old'}; Error.captureStackTrace(target);`, "target shape"},
		{`const target={get stack():string {return 'old';}}; Error.captureStackTrace(target);`, "target shape"},
		{`import type {Stats} from 'node:fs'; const error=new Error(); console.log('stack' in error ? 'yes' : 'no');`, "stack presence receiver"},
		{`import type {Stats} from 'node:fs'; const error=new Error(); const {stack}=error; console.log(typeof stack);`, "stack destructuring"},
		{`import type {Stats} from 'node:fs'; const error=new Error(); read(); Error.captureStackTrace(error); function read():void { console.log(typeof error.stack); }`, "without a preceding captureStackTrace"},
		{`const source={stack:3}; const target:{}=source; Error.captureStackTrace(target);`, "target shape"},
		{`Error.stackTraceLimit += 1;`, "evaluation-order proof"},
		{`import type {Stats} from 'node:fs'; const error=new Error('uncaptured'); const view:{stack?:string}=error; console.log(typeof view.stack);`, "known plain object"},
		{`Error.captureStackTrace([1]);`, "present native object"},
		{`const target={stack:3}; Error.captureStackTrace(target);`, "non-string stack"},
		{`const target={}; Object.freeze(target); Error.captureStackTrace(target);`, "freezes objects"},
		{`import type {Stats} from 'node:fs'; const error=new Error('uncaptured'); console.log(typeof error.stack);`, "without a preceding captureStackTrace"},
		{`import type {Stats} from 'node:fs'; const error=new Error('uncaptured'); console.log(Object.hasOwn(error,'stack')?'yes':'no');`, "without a preceding captureStackTrace"},
	} {
		t.Run(probe.reason, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSource(t, probe.source)
			var notYet *NotYet
			if !errors.As(err, &notYet) || !strings.Contains(notYet.What, probe.reason) {
				t.Fatalf("want named NotYet %s: %v", probe.reason, err)
			}
		})
	}
}

func TestOrdinaryNonStringStackReads(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		`class Parser { stack:number[]=[]; peek():number { return this.stack[this.stack.length-1] ?? -1; } } const parser=new Parser(); parser.stack.push(7); console.log(parser.peek().toString());`,
		`class Parser { stack:number=7; peek():number { return this.stack; } } const parser=new Parser(); console.log(parser.peek().toString());`,
	} {
		t.Run(source, func(t *testing.T) {
			t.Parallel()
			if _, err := lowerSource(t, source); err != nil {
				t.Fatal(err)
			}
		})
	}
}
