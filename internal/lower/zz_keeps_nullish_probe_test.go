package lower

import (
	"fmt"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

// Proof-only (#qxt10dd scout's find, never merged): does a cast that keeps undefined, whose present pair takes the
// untagged arm or the class road, lower with its check, or as the bare operand?
func TestZZKeepsNullishUntaggedAndClassProbe(t *testing.T) {
	for _, probe := range []struct{ name, source string }{
		{"untagged", `interface A { readonly a: number; readonly n: string }
interface B { readonly b: string }
function f(x: A | B | undefined): A | undefined { return x as A | undefined; }
const b: B = { b: "bee" };
const got = f(b);
console.log(got === undefined ? "undefined" : String(got.a) + got.n);
`},
		{"class", `class Base { constructor(readonly x: number) {} }
class Sub extends Base { constructor(x: number, readonly y: string) { super(x); } }
function g(v: Base | undefined): Sub | undefined { return v as Sub | undefined; }
const got = g(new Base(1));
console.log(got === undefined ? "undefined" : String(got.y.length));
`},
		{"untagged control, no undefined", `interface A { readonly a: number; readonly n: string }
interface B { readonly b: string }
function f(x: A | B): A { return x as A; }
const b: B = { b: "bee" };
const got = f(b);
console.log(String(got.a) + got.n);
`},
	} {
		program, err := lowerSourceAs(t, probe.source, ".ts")
		if err != nil {
			fmt.Printf("PROBE %s: lowering stops: %v\n", probe.name, err)
			continue
		}
		checks := 0
		views := 0
		for _, function := range program.Functions {
			walk(function.Body, func(node any) bool {
				switch value := node.(type) {
				case ir.CheckedCast:
					checks++
				case ir.Property:
					if value.View != "" {
						views++
					}
				}
				return true
			})
		}
		code, stdout, stderr := brandedNative(t, program)
		fmt.Printf("PROBE %s: lowers; checked casts %d, viewed reads %d; native exit %d stdout %q stderr %q\n", probe.name, checks, views, code, strings.TrimSpace(stdout), strings.TrimSpace(stderr))
	}
}
