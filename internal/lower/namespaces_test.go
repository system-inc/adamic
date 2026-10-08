package lower

import (
	"errors"
	"os"
	"strings"
	"testing"
)

func TestNamespaceLimitsStayLoud(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source, reason string }{
		{"computed enum", "namespace N {enum E {A=Math.floor(2)}}", "computed enum"},
		{"reopened enum", "namespace N {enum E {A} enum E {B=1}}", "merged enum"},
		{"early enum", "namespace N {function read():number{return E.A;} const x=read(); enum E {A}}", "enum before"},
		{"mutable object escape", "namespace N {export let x=1;} const value=N; value.x=2;", "namespace object"},
		{"object value", "namespace N {export const x=1;} const value=N;", "namespace object"},
		{"reflection", "namespace N {export const x=1;} console.log(Object.keys(N).join(' '));", "namespace object"},
		{"reopening", "namespace N {export const x=1;} namespace N {export const y=2;}", "reopened namespace"},
		{"class merge", "class N {} namespace N {export const x=1;}", "namespace merged with a class"},
		{"nested object var", "namespace N {var {inner:{x}}={inner:{x:1}};}", "destructuring inside a namespace"},
		{"rest object var", "namespace N {var {x,...rest}={x:1,y:2};}", "destructuring inside a namespace"},
		{"default object var", "namespace N {var {x=1}={x:2};}", "destructuring inside a namespace"},
		{"destructured state", "namespace N {const [x]=[1];}", "destructuring inside a namespace"},
		{"computed member", "namespace N {export const x=1;} console.log(`${N['x']}`);", "namespace object"},
		{"replace function", "namespace N {export function read():number{return 1;}} N.read=()=>2;", "replacing a namespace export"},
		{"early call", "function early():number{return N.x;} const x=early(); namespace N {export const x=1;}", "namespace read before runtime initialization"},
		{"early read", "const x=N.read; namespace N {export function read():number{return 1;}}", "namespace read before"},
		{"block var", "namespace N {if(true){var x=1;}}", "var inside namespace control flow"},
		{"nested early read", "namespace N {const x=Inner.read; export namespace Inner {export function read():number{return 1;}}}", "namespace read before"},
		{"explicit receiver", "namespace N {export const x=1; export function read(this:{readonly x:number}):number{return 1;}}", "explicit namespace-function"},
		{"ambient", "declare namespace N {export interface T {readonly x:number;}}", "ambient namespace"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			_, err := lowerSource(t, probe.source)
			var notYet *NotYet
			if !errors.As(err, &notYet) || !strings.Contains(err.Error(), probe.reason) {
				t.Fatalf("got %v, want NotYet %s", err, probe.reason)
			}
		})
	}
}

func TestNamespaceReceiverRefusal(t *testing.T) {
	t.Parallel()
	_, err := lowerSource(t, "namespace N {export const x=1; export function read(this:{readonly x:number}|void):number{return this?.x ?? 0;} export function detached():number{return read();}}")
	var refused *Refused
	if !errors.As(err, &refused) || !strings.Contains(err.Error(), "different receivers") {
		t.Fatalf("got %v", err)
	}
}

func TestNamespaceTypesErase(t *testing.T) {
	t.Parallel()
	program, err := lowerSource(t, "namespace N {export interface Box {readonly x:number;} export namespace Inner {export type Number=number;}} const value:N.Box={x:7}; const x:N.Inner.Number=value.x; console.log(`${x}`);")
	if err != nil {
		t.Fatal(err)
	}
	for _, local := range program.Locals {
		if local.Name == "N" || local.Name == "Inner" {
			t.Fatal("type namespace made a binding")
		}
	}
}

func TestTracingNamespaceEscapeStaysNotYet(t *testing.T) {
	source, err := os.ReadFile("testdata/namespaces_notyet/tracing_escape.a")
	if err != nil {
		t.Fatal(err)
	}
	_, err = lowerSource(t, string(source))
	var notYet *NotYet
	if !errors.As(err, &notYet) || !strings.Contains(err.Error(), "no runtime container is emitted") {
		t.Fatalf("got %v, want the explicit missing runtime-container reason", err)
	}
}

func TestDebugNamespaceMergesStayNotYet(t *testing.T) {
	for _, test := range []struct{ name, reason string }{
		{"class_merge.a", "constructor identity"},
	} {
		t.Run(test.name, func(t *testing.T) {
			source, err := os.ReadFile("testdata/namespaces_notyet/" + test.name)
			if err != nil {
				t.Fatal(err)
			}
			_, err = lowerSource(t, string(source))
			var notYet *NotYet
			if !errors.As(err, &notYet) || !strings.Contains(err.Error(), test.reason) {
				t.Fatalf("got %v, want NotYet %s", err, test.reason)
			}
		})
	}
}

func TestNamespaceReturnedAssignmentLimits(t *testing.T) {
	for _, source := range []string{
		"namespace N {let text=''; export function set():string{return text='built'.repeat(2);}}",
		"namespace N {let optional:number|undefined; export function set():number{return optional=1;}}",
	} {
		_, err := lowerSource(t, source)
		var notYet *NotYet
		if !errors.As(err, &notYet) || !strings.Contains(err.Error(), "only scalar singleton assignment is proven") {
			t.Fatalf("got %v, want the scalar returned-assignment boundary", err)
		}
	}
}

func TestNamespaceInitializationReachability(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"direct", "helper", "cycle"} {
		t.Run(name, func(t *testing.T) {
			source, err := os.ReadFile("testdata/namespaces_notyet/reaching_" + name + ".a")
			if err != nil {
				t.Fatal(err)
			}
			_, err = lowerSource(t, string(source))
			var notYet *NotYet
			if !errors.As(err, &notYet) || !strings.Contains(err.Error(), "through a reachable call") {
				t.Fatalf("reaching call lost its initialization refusal: %v", err)
			}
		})
	}
}

func TestNamespaceClosedCallGraphEdges(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"function read():boolean{return N.x;} class C {static readonly x=read();} namespace N {export let x=false;}",
		"function read():boolean{return N.x;} const alias=read; alias(); namespace N {export let x=false;}",
		"function read():boolean{return N.x;} [1].map(read); namespace N {export let x=false;}",
		"function read(x:boolean=N.x):boolean{return x;} read(); namespace N {export let x=false;}",
		"function one(x:boolean):boolean{return x ? two(false) : N.x;} function two(x:boolean):boolean{return one(x);} one(true); namespace N {export let x=false;}",
	} {
		_, err := lowerSource(t, source)
		var notYet *NotYet
		if !errors.As(err, &notYet) || !strings.Contains(err.Error(), "through a reachable call") {
			t.Fatalf("got %v", err)
		}
	}
}

func TestEmptyNeverMapCannotGainWritableInhabitants(t *testing.T) {
	t.Parallel()
	_, err := lowerSource(t, "const empty = new Map<never, never>(); const wide: Map<string, number> = empty; wide.set('x', 1);")
	var refused *Refused
	if !errors.As(err, &refused) {
		t.Fatalf("got %v, want invariant mutable view refused", err)
	}
}

func TestDebugNamespaceObservationBoundary(t *testing.T) {
	t.Parallel()
	source, err := os.ReadFile("../../stage3/namespaces/debug-groups/object_observation.a")
	if err != nil {
		t.Fatal(err)
	}
	_, err = lowerSource(t, string(source))
	var notYet *NotYet
	if !errors.As(err, &notYet) || !strings.Contains(err.Error(), "no runtime container is emitted") {
		t.Fatalf("Debug container boundary lost: %v", err)
	}
}

func TestDebugUnknownAssertionBoundaries(t *testing.T) {
	t.Parallel()
	for _, test := range []struct{ name, source, reason string }{
		{"unknown condition", "namespace Debug {export function assert(expression:unknown):asserts expression {if(!expression){throw new Error('False expression.');}}} Debug.assert(true);", "asserts cond needs a boolean parameter"},
		{"empty generic assertion", "namespace Debug {export function type<T>(value:unknown):asserts value is T {}} function use(value:unknown):number {Debug.type<number>(value);return value+1;} console.log(`${use('wrong')}`);", "normal return has not narrowed"},
	} {
		t.Run(test.name, func(t *testing.T) {
			_, err := lowerSource(t, test.source)
			var refused *Refused
			if !errors.As(err, &refused) || !strings.Contains(err.Error(), test.reason) {
				t.Fatalf("Debug assertion boundary lost: %v", err)
			}
		})
	}
}

func TestDebugLocalConstEnums(t *testing.T) {
	t.Parallel()
	source, err := os.ReadFile("../../stage3/namespaces/debug-groups/local_const_enums.a")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = lowerSource(t, string(source)); err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct{ source, reason string }{
		{"function run():number {const enum E {A=1} const read=():number=>E.A;return read();} console.log(`${run()}`);", "deferred body"},
		{"function run():number {enum E {A=1} return E.A;}", "enum inside a function"},
	} {
		_, err := lowerSource(t, test.source)
		var notYet *NotYet
		if !errors.As(err, &notYet) || !strings.Contains(err.Error(), test.reason) {
			t.Fatalf("local enum boundary lost: %v", err)
		}
	}
}

func TestNamespaceNestedReceiver(t *testing.T) {
	source, err := os.ReadFile("../../stage3/namespaces/debug-groups/native-namespace-object-receiver.a")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := lowerSource(t, string(source)); err != nil {
		t.Fatal(err)
	}
}

func TestNamespaceReceiverLimits(t *testing.T) {
	for _, test := range []struct{ source, reason string }{
		{"namespace N {export const x=1;export function read(this:{x:number}):number{return this.x;}} const detached=N.read;", "closed-world receiver"},
		{"namespace N {export const x=1;export function read(this:{x:number}):{x:number}{return this;}} console.log(`${N.read().x}`);", "receiver object"},
		{"namespace N {export let x=1;export function write(this:{x:number}):void{this.x=2;}} N.write();", "write through a namespace receiver"},
	} {
		_, err := lowerSource(t, test.source)
		var ny *NotYet
		if !errors.As(err, &ny) || !strings.Contains(err.Error(), test.reason) {
			t.Fatalf("receiver boundary lost: %v", err)
		}
	}
}

func TestCallableNamespaceBoundaries(t *testing.T) {
	for _, source := range []string{
		"function log():void{} namespace log {export const level=1;} const escaped=log;",
		"function log():void{} namespace log {export const level=1;} console.log(log.name);",
		"function log():void{} namespace log {export const level=1;} console.log(Object.keys(log).join(','));",
	} {
		_, err := lowerSource(t, source)
		var ny *NotYet
		if !errors.As(err, &ny) || !strings.Contains(err.Error(), "callable namespace object") {
			t.Fatalf("callable object boundary lost: %v", err)
		}
	}
}
