package lower

import (
	"errors"
	"strings"
	"testing"
)

func TestCallableNamespaceLimitsStayLoud(t *testing.T) {
	for _, probe := range []struct{ name, source, reason string }{
		{"escape", "function f(){} namespace f {export const x=1;} const alias=f;", "callable namespace object"},
		{"reflection", "function f(){} namespace f {export const x=1;} console.log(Object.keys(f).join(' '));", "callable namespace object"},
		{"replacement", "function f(){} namespace f {export function g(){}} f.g=()=>{};", "replacing a namespace export"},
		{"intrinsic", "function f(){} namespace f {export const name='shadow';}", "function intrinsic"},
		{"intrinsic method", "function f(){} namespace f {export function call(){}}", "function intrinsic"},
		{"early", "function f(){} namespace f {export const x=1;} function h(){return f.x;} const value=h(); namespace Later {export const y=2;}", "call before all runtime namespaces"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			_, err := lowerSource(t, probe.source)
			var notYet *NotYet
			if !errors.As(err, &notYet) || !strings.Contains(err.Error(), probe.reason) {
				t.Fatalf("got %v, want %s", err, probe.reason)
			}
		})
	}
}

func TestCallableNamespaceReceiverStaysLoud(t *testing.T) {
	_, err := lowerSource(t, "function f(this:{x:number}):number{return this.x;} namespace f {export const x=1;}")
	var refused *Refused
	if !errors.As(err, &refused) || !strings.Contains(err.Error(), "different receivers") {
		t.Fatalf("got %v", err)
	}
}
