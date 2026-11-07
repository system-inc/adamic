package lower

import (
	"errors"
	"os"
	"strings"
	"testing"
)

func TestNamespaceLimitsStayLoud(test *testing.T) {
	test.Parallel()
	for _, probe := range []struct{ name, source, reason string }{
		{"computed enum", "namespace N {enum E {A=Math.floor(2)}}", "computed enum"},
		{"reopened enum", "namespace N {enum E {A} enum E {B=1}}", "merged enum"},
		{"early enum", "namespace N {function read():number{return E.A;} const x=read(); enum E {A}}", "enum before"},
		{"mutable object escape", "namespace N {export let x=1;} const value=N; value.x=2;", "namespace object"},
		{"object value", "namespace N {export const x=1;} const value=N;", "namespace object"},
		{"reflection", "namespace N {export const x=1;} console.log(Object.keys(N).join(' '));", "namespace object"},
		{"reopening", "namespace N {export const x=1;} namespace N {export const y=2;}", "reopened namespace"},
		{
			"function merge escape",
			"function N():number{return 1;} namespace N {export const x=1;} const alias=N;",
			"callable namespace object",
		},
		{"class merge", "class N {} namespace N {export const x=1;}", "merged with a runtime value"},
		{"destructured state", "namespace N {const [x]=[1];}", "destructuring inside a namespace"},
		{"nested object var", "namespace N {var {inner:{x}}={inner:{x:1}};}", "destructuring inside a namespace"},
		{"rest object var", "namespace N {var {x,...rest}={x:1,y:2};}", "destructuring inside a namespace"},
		{"default object var", "namespace N {var {x=1}={x:2};}", "destructuring inside a namespace"},
		{"computed member", "namespace N {export const x=1;} console.log(`${N['x']}`);", "namespace object"},
		{
			"replace function",
			"namespace N {export function read():number{return 1;}} N.read=()=>2;",
			"replacing a namespace export",
		},
		{
			"early call",
			"function early():number{return N.x;} const x=early(); namespace N {export const x=1;}",
			"namespace read before runtime initialization",
		},
		{
			"early read",
			"const x=N.read; namespace N {export function read():number{return 1;}}",
			"namespace read before",
		},
		{"block var", "namespace N {if(true){var x=1;}}", "var inside namespace control flow"},
		{
			"nested early read",
			"namespace N {const x=Inner.read; export namespace Inner {export function read():number{return 1;}}}",
			"namespace read before",
		},
		{
			"explicit receiver",
			"namespace N {export const x=1; export function read(this:{readonly x:number}):number{return 1;}}",
			"explicit namespace-function",
		},
	} {
		test.Run(probe.name, func(test *testing.T) {
			_, err := lowerSource(test, probe.source)
			var notYet *NotYet
			if !errors.As(err, &notYet) || !strings.Contains(err.Error(), probe.reason) {
				test.Fatalf("got %v, want NotYet %s", err, probe.reason)
			}
		})
	}
}

func TestNamespaceReceiverRefusal(test *testing.T) {
	test.Parallel()
	_, err := lowerSource(test, "namespace N {export function read(this:{readonly x:number}):number{return this.x;}}")
	var refused *Refused
	if !errors.As(err, &refused) || !strings.Contains(err.Error(), "different receivers") {
		test.Fatalf("got %v", err)
	}
}

func TestNamespaceTypesErase(test *testing.T) {
	test.Parallel()
	// Behavior cannot observe erased type-only namespace bindings.
	program := lowersAndAgreesWithNode(
		test,
		"namespace N {export interface Box {readonly x:number;} export namespace Inner {export type Number=number;}} const value:N.Box={x:7}; const x:N.Inner.Number=value.x; console.log(`${x}`);",
	)
	for _, local := range program.Locals {
		if local.Name == "N" || local.Name == "Inner" {
			test.Fatal("type namespace made a binding")
		}
	}
}

func TestTracingNamespaceEscapeStaysNotYet(test *testing.T) {
	test.Parallel()
	source, err := os.ReadFile("testdata/namespaces_notyet/tracing_escape.a")
	if err != nil {
		test.Fatal(err)
	}
	_, err = lowerSource(test, string(source))
	var notYet *NotYet
	if !errors.As(err, &notYet) || !strings.Contains(err.Error(), "no runtime container is emitted") {
		test.Fatalf("got %v, want the explicit missing runtime-container reason", err)
	}
}

func TestDebugNamespaceMergedCapabilities(test *testing.T) {
	test.Parallel()
	for _, testCase := range []struct{ name, reason string }{
		{"debug_log.a", ""},
		{"class_merge.a", "merged with a runtime value"},
		{"debug_class.a", ""},
	} {
		test.Run(testCase.name, func(test *testing.T) {
			test.Parallel()
			source, err := os.ReadFile("testdata/namespaces_notyet/" + testCase.name)
			if err != nil {
				test.Fatal(err)
			}
			if testCase.reason == "" {
				lowersAndAgreesWithNode(test, string(source))
				return
			}
			_, err = lowerSource(test, string(source))
			var notYet *NotYet
			if !errors.As(err, &notYet) || !strings.Contains(err.Error(), testCase.reason) {
				test.Fatalf("got %v, want NotYet %s", err, testCase.reason)
			}
		})
	}
}

func TestNamespaceReturnedAssignmentLimits(test *testing.T) {
	test.Parallel()
	for _, source := range []string{
		"namespace N {let text=''; export function set():string{return text='built'.repeat(2);}}",
		"namespace N {let optional:number|undefined; export function set():number{return optional=1;}}",
	} {
		_, err := lowerSource(test, source)
		var notYet *NotYet
		if !errors.As(err, &notYet) || !strings.Contains(err.Error(), "only scalar singleton assignment is proven") {
			test.Fatalf("got %v, want the scalar returned-assignment boundary", err)
		}
	}
}

func TestNamespaceInitializationReachability(test *testing.T) {
	test.Parallel()
	for _, name := range []string{"direct", "helper", "cycle"} {
		test.Run(name, func(test *testing.T) {
			source, err := os.ReadFile("testdata/namespaces_notyet/reaching_" + name + ".a")
			if err != nil {
				test.Fatal(err)
			}
			_, err = lowerSource(test, string(source))
			var notYet *NotYet
			if !errors.As(err, &notYet) || !strings.Contains(err.Error(), "through a reachable call") {
				test.Fatalf("reaching call lost its initialization refusal: %v", err)
			}
		})
	}
}

func TestNamespaceClosedCallGraphEdges(test *testing.T) {
	test.Parallel()
	for _, source := range []string{
		"function read():boolean{return N.x;} class C {static readonly x=read();} namespace N {export let x=false;}",
		"function read():boolean{return N.x;} const alias=read; alias(); namespace N {export let x=false;}",
		"function read():boolean{return N.x;} [1].map(read); namespace N {export let x=false;}",
		"function read(x:boolean=N.x):boolean{return x;} read(); namespace N {export let x=false;}",
		"function one(x:boolean):boolean{return x ? two(false) : N.x;} function two(x:boolean):boolean{return one(x);} one(true); namespace N {export let x=false;}",
	} {
		_, err := lowerSource(test, source)
		var notYet *NotYet
		if !errors.As(err, &notYet) || !strings.Contains(err.Error(), "through a reachable call") {
			test.Fatalf("got %v", err)
		}
	}
}

func TestEmptyNeverMapCannotGainWritableInhabitants(test *testing.T) {
	test.Parallel()
	_, err := lowerSource(
		test,
		"const empty = new Map<never, never>(); const wide: Map<string, number> = empty; wide.set('x', 1);",
	)
	var refused *Refused
	if !errors.As(err, &refused) {
		test.Fatalf("got %v, want invariant mutable view refused", err)
	}
}
