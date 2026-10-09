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
		{"function merge escape", "function N():number{return 1;} namespace N {export const x=1;} const alias=N;", "callable namespace object"},
		{"class merge", "class N {} namespace N {export const x=1;}", "merged with a runtime value"},
		{"destructured state", "namespace N {const [x]=[1];}", "destructuring inside a namespace"},
		{"nested object var", "namespace N {var {inner:{x}}={inner:{x:1}};}", "destructuring inside a namespace"},
		{"rest object var", "namespace N {var {x,...rest}={x:1,y:2};}", "destructuring inside a namespace"},
		{"default object var", "namespace N {var {x=1}={x:2};}", "destructuring inside a namespace"},
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
	_, err := lowerSource(t, "namespace N {export function read(this:{readonly x:number}):number{return this.x;}}")
	var refused *Refused
	if !errors.As(err, &refused) || !strings.Contains(err.Error(), "different receivers") {
		t.Fatalf("got %v", err)
	}
}

func TestNamespaceTypesErase(t *testing.T) {
	t.Parallel()
	// Behavior cannot observe erased type-only namespace bindings.
	program := lowersAndAgreesWithNode(t, "namespace N {export interface Box {readonly x:number;} export namespace Inner {export type Number=number;}} const value:N.Box={x:7}; const x:N.Inner.Number=value.x; console.log(`${x}`);")
	for _, local := range program.Locals {
		if local.Name == "N" || local.Name == "Inner" {
			t.Fatal("type namespace made a binding")
		}
	}
}

func TestTracingNamespaceEscapeLowers(t *testing.T) {
	t.Parallel()
	source, err := os.ReadFile("testdata/namespaces_notyet/tracing_escape.a")
	if err != nil {
		t.Fatal(err)
	}
	_, err = lowerSource(t, string(source))
	if err != nil {
		t.Fatal(err)
	}
}

func TestDebugNamespaceMergedCapabilities(t *testing.T) {
	t.Parallel()
	for _, test := range []struct{ name, reason string }{
		{"debug_log.a", ""},
		{"class_merge.a", "merged with a runtime value"},
		{"debug_class.a", ""},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			source, err := os.ReadFile("testdata/namespaces_notyet/" + test.name)
			if err != nil {
				t.Fatal(err)
			}
			if test.reason == "" {
				lowersAndAgreesWithNode(t, string(source))
				return
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
	t.Parallel()
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
