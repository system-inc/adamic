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
		{"function merge", "function N():number{return 1;} namespace N {export const x=1;}", "namespace merged with a function"},
		{"class merge", "class N {} namespace N {export const x=1;}", "namespace merged with a class"},
		{"nested object var", "namespace N {var {inner:{x}}={inner:{x:1}};}", "destructuring inside a namespace"},
		{"rest object var", "namespace N {var {x,...rest}={x:1,y:2};}", "destructuring inside a namespace"},
		{"default object var", "namespace N {var {x=1}={x:2};}", "destructuring inside a namespace"},
		{"destructured state", "namespace N {const [x]=[1];}", "destructuring inside a namespace"},
		{"computed member", "namespace N {export const x=1;} console.log(`${N['x']}`);", "namespace object"},
		{"replace function", "namespace N {export function read():number{return 1;}} N.read=()=>2;", "replacing a namespace export"},
		{"early call", "function early():number{return N.x;} const x=early(); namespace N {export const x=1;}", "call before all runtime namespaces"},
		{"early read", "const x=N.read; namespace N {export function read():number{return 1;}}", "namespace read before"},
		{"initializer call", "function get():number{return 1;} namespace N {export const x=get();}", "call before all runtime namespaces"},
		{"unknown body call", "function effect():void{} namespace N {effect();}", "call before all runtime namespaces"},
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
		{"debug_log.a", "callable object properties"},
		{"class_merge.a", "constructor identity"},
		{"debug_class.a", "constructor registration"},
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
