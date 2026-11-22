package lower

import (
	"errors"
	"strings"
	"testing"
)

// valueIdentityVisitor is an interface of two generic call signatures, as tsc's NodeVisitor is, and an
// overload set that can be one, which stays NotYet where it's handed over as a value (a polymorphic
// function value needs dispatch, Adamic's ruling on #3zvr9mq).
const valueIdentityVisitor = `interface Visitor<A> {
	<T extends number>(x: T, tag: A): T;
	<T extends string>(x: T, tag: A): T;
}
function id<T extends number>(x: T, tag: number): T;
function id<T extends string>(x: T, tag: number): T;
function id(x: number | string, tag: number): number | string { return x; }
function apply(v: Visitor<number>): number { return v(3, 1); }
`

// A value whose type is its slot's own binds nothing at the value (#3zvr9mq, tsc's 22 visitor forwards): a
// Visitor<number> parameter forwarded into a Visitor<number> parameter passes, so the program's first stop
// is the overload set handed over as a value on the line below it (10). A Visitor<number> forwarded into a
// Visitor<1> parameter, types that differ by a type argument, still stops at the forward (10), where its
// binding can't be read. Mutant: the identity test dropped from valueTypeArgumentBindings, and the forward
// row stops at its forward, line 9.
func TestValueOfItsSlotsTypeBindsNothing(t *testing.T) {
	t.Parallel()
	for _, row := range []struct {
		name, source, at string
	}{
		{"forward", valueIdentityVisitor + "function forward(v: Visitor<number>): number { return apply(v); }\nconst made: Visitor<number> = id;\nconsole.log(String(forward(made)));\n", "main.ts:10:"},
		{"forward_differs", valueIdentityVisitor + "function applyOne(v: Visitor<1>): number { return v(3, 1); }\nfunction forward(v: Visitor<number>): number { return applyOne(v); }\nconst made: Visitor<number> = id;\nconsole.log(String(forward(made)));\n", "main.ts:10:"},
	} {
		t.Run(row.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSourceAs(t, row.source, ".ts")
			var notYet *NotYet
			if !errors.As(err, &notYet) {
				t.Fatalf("got %v, want NotYet at %s", err, row.at)
			}
			t.Logf("%s: %s: %s", row.name, notYet.Where, notYet.What)
			if !strings.Contains(notYet.Where, row.at) {
				t.Fatalf("NotYet at %s: %q, want it at %s", notYet.Where, notYet.What, row.at)
			}
		})
	}
}
