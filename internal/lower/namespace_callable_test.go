package lower

import (
	"errors"
	"strings"
	"testing"
)

func TestCallableNamespaceLimitsStayLoud(test *testing.T) {
	test.Parallel()
	for _, probe := range []struct{ name, source, reason string }{
		{"escape", "function f(){} namespace f {export const x=1;} const alias=f;", "callable namespace object"},
		{
			"reflection",
			"function f(){} namespace f {export const x=1;} console.log(Object.keys(f).join(' '));",
			"callable namespace object",
		},
		{
			"replacement",
			"function f(){} namespace f {export function g(){}} f.g=()=>{};",
			"replacing a namespace export",
		},
		{"intrinsic", "function f(){} namespace f {export const name='shadow';}", "function intrinsic"},
		{"intrinsic method", "function f(){} namespace f {export function call(){}}", "function intrinsic"},
		{
			"early",
			"function f(){} function h(){return f.x;} const value=h(); namespace f {export const x=1;}",
			"namespace read before runtime initialization",
		},
	} {
		test.Run(probe.name, func(test *testing.T) {
			test.Parallel()
			_, lowerSourceError := lowerSource(test, probe.source)
			var notYet *NotYet
			if !errors.As(lowerSourceError, &notYet) || !strings.Contains(lowerSourceError.Error(), probe.reason) {
				test.Fatalf("got %v, want %s", lowerSourceError, probe.reason)
			}
		})
	}
}

func TestCallableNamespaceReceiverStaysLoud(test *testing.T) {
	test.Parallel()
	_, lowerSourceError := lowerSource(
		test,
		"function f(this:{x:number}):number{return this.x;} namespace f {export const x=1;}",
	)
	var refused *Refused
	if !errors.As(lowerSourceError, &refused) || !strings.Contains(lowerSourceError.Error(), "different receivers") {
		test.Fatalf("got %v", lowerSourceError)
	}
}
