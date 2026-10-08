package lower

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/system-inc/adamic/internal/load"
	"strings"
	"testing"
)

func TestErrorCaptureBoundaries(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ source, reason string }{
		{`const target={stack:'old'}; Error.captureStackTrace(target);`, "target shape"},
		{`const target={get stack():string {return 'old';}}; Error.captureStackTrace(target);`, "target shape"},
		{`import type {Stats} from 'node:fs'; const error=new Error(); console.log('stack' in error ? 'yes' : 'no');`, "stack presence receiver"},
		{`import type {Stats} from 'node:fs'; const error=new Error(); const {stack}=error; console.log(typeof stack);`, "Error destructuring"},
		{`const source={stack:3}; const target:{}=source; Error.captureStackTrace(target);`, "target shape"},
		{`Error.stackTraceLimit += 1;`, "evaluation-order proof"},
		{`import type {Stats} from 'node:fs'; const error=new Error('uncaptured'); const view:{stack?:string}=error; console.log(typeof view.stack);`, "known plain object"},
		{`Error.captureStackTrace([1]);`, "present native object"},
		{`const target={stack:3}; Error.captureStackTrace(target);`, "non-string stack"},
		{`const target={}; Object.freeze(target); Error.captureStackTrace(target);`, "freezes objects"},
		{`const target={}; Object.seal(target); Error.captureStackTrace(target);`, "sealed/nonextensible"},
		{`const target={}; Object.preventExtensions(target); Error.captureStackTrace(target);`, "sealed/nonextensible"},
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

func TestErrorObjectBoundaries(t *testing.T) {
	for _, probe := range []struct{ source, reason string }{
		{`function make(message?:string):Error {return new Error(message);} const e=make();`, "message-presence"},
		{`const options:ErrorOptions={cause:3}; const error=new Error('x',options);`, "plain literal"},
		{`const error=new Error('x',{get cause():number {return 3}});`, "HasProperty/Get"},
		{`const error=new AggregateError(new Set<number>([1]),'many');`, "iterable protocols"},
		{`const error=new Error(); error.message += 'x';`, "evaluation-order"},
		{`const error=new Error(); error.cause = 3;`, "boxed-property"},
		{`const error=new AggregateError([1]); error.errors = [2];`, "boxed-property"},
		{`const error=new Error(); console.log(error?.message ?? 'missing');`, "short-circuit"},
		{`const fake:Error={name:'fake',message:'x'}; console.log(fake.toString());`, "non-Error object"},
		{`const error=new Error(); const copy={...error};`, "enumerable-descriptor"},
		{`import type {Stats} from 'node:fs'; function stack(e:Error):string {return e.stack ?? '';}`, "metadata proof"},
		{`import type {Stats} from 'node:fs'; function own(e:Error):boolean {return Object.hasOwn(e,'name');}`, "metadata proof"},
		{`const e=new Error(); Object.assign(e,{name:'renamed'});`, "own-property metadata"},
		{`const e=new Error(); e['name']='renamed';`, "bracket writes"},
		{`const e=new Error(); console.log(e['message']);`, "bracket members"},
		{`const e=new Error(); const {message}=e;`, "destructuring"},
		{`import {readFileSync} from 'node:fs'; Error.prototype.name='Renamed';`, "host Error ancestry"},
		{`const target:{name:string|null}={name:null}; Error.captureStackTrace(target);`, "nullable field"},
		{`const proto=Error.prototype; Object.getPrototypeOf(Object.getPrototypeOf(Object.getPrototypeOf(proto)));`, "getPrototypeOf(null)"},
		{`function walk(e:Error):void {Object.getPrototypeOf(Object.getPrototypeOf(Object.getPrototypeOf(e)));} walk(Error.prototype);`, "getPrototypeOf(null)"},
		{`const proto=Object.prototype;`, "generic prototype"},
		{`const plain={value:1}; Object.freeze(plain); const e=new Error(); e.name='changed';`, "descriptor failures"},
		{`const target={name:3}; Error.captureStackTrace(target);`, "ToPrimitive"},
		{`Object.getPrototypeOf(Object.getPrototypeOf(Object.getPrototypeOf(Error.prototype)));`, "getPrototypeOf(null)"},
	} {
		t.Run(probe.reason, func(t *testing.T) {
			_, err := lowerSource(t, probe.source)
			var notYet *NotYet
			if !errors.As(err, &notYet) || !strings.Contains(notYet.What, probe.reason) {
				t.Fatalf("want named refusal %q: %v", probe.reason, err)
			}
		})
	}
}

func TestTscErrorCompilerBoundaries(t *testing.T) {
	// debug.ts:199 keeps the source's truthy message test. It is not replaced by
	// a length comparison just to make the library fixture pass.
	_, err := lowerSource(t, `function fail(message?:string):void {const e=new Error(message ? "Debug Failure. " + message : "Debug Failure."); throw e;} fail();`)
	if err == nil || !strings.Contains(err.Error(), "string as a condition") {
		t.Fatalf("want compiler string-truthiness refusal: %v", err)
	}
	// debug.ts:203: stored Error throw remains independently compiler-owned.
	_, err = lowerSource(t, `const e=new Error('Debug Failure.'); throw e;`)
	if err == nil || !strings.Contains(err.Error(), "isn't made where it's thrown") {
		t.Fatalf("want compiler stored-throw refusal: %v", err)
	}
}

func TestTscErrorSourceShapes(t *testing.T) {
	_, err := lowerSource(t, "function make(moduleName:string,initialDir:string,failedLookupLocations?:string[]):Error {return new Error(`Could not resolve JS module '${moduleName}' starting at '${initialDir}'. Looked in: ${failedLookupLocations?.join(\", \")}`);} const e=make('m','d'); console.log(e.message);")
	if err == nil || !strings.Contains(err.Error(), "optional call") {
		t.Fatalf("moduleNameResolver.ts:1681 compiler optional-call refusal: %v", err)
	}
	// tracing.ts:66 uses the unmodified catch expression. The stock embedded
	// checker rejects its unknown catch variable before library lowering.
	path := filepath.Join(t.TempDir(), "main.a")
	source := "try {throw new Error('fs');} catch(e) {const failure=new Error(`tracing requires having fs\\n(original error: ${e.message || e})`); console.log(failure.message);}"
	if err := os.WriteFile(path, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
	_, err = load.Load([]string{path})
	if err == nil || !strings.Contains(err.Error(), "TS18046") {
		t.Fatalf("tracing.ts:66 checker catch-variable refusal: %v", err)
	}
}
