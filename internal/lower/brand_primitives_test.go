package lower

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/load"
)

func TestBrandPrimitivesAgreeWithNode(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, value, primitive, member string }{
		{"string", "'abc'", "string", "String(b.length)"},
		{"number", "3", "number", "b.toFixed(0)"},
		{"boolean", "true", "boolean", "String(b.valueOf())"},
	} {
		for _, field := range []string{"readonly __b?: undefined", "readonly __b: void"} {
			t.Run(probe.name+"/"+field, func(t *testing.T) {
				t.Parallel()
				source := "const b = " + probe.value + " as " + probe.primitive + " & { " + field + " };\nconsole.log(String(b.__b));\nconsole.log(" + probe.member + ");\n"
				lowersAndAgreesWithNode(t, source)
			})
		}
	}
}

// Adamic ruled bigint and symbol not brandable yet on #yth65gz (Nov 10).
// Mutant run and restored: admitting BigIntLike in phantomParts breaks the
// bigint read-proof and required-void read pins below. This earlier stop stays loud.
func TestBrandPrimitivesNotYet(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source, what string }{
		{"bigint", "const b = 1n as bigint & { readonly __b?: undefined }; console.log(String(b + 1n)); console.log(String(b.__b));", "bigint & { readonly __b?: undefined; }"},
		{"symbol", "const s = Symbol('x') as symbol & { readonly __b?: undefined }; console.log(String(s.description)); console.log(String(s.__b));", "symbol & { readonly __b?: undefined; }"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSourceAs(t, probe.source, ".ts")
			var gap *NotYet
			want := "main.ts:1:7: stage 0 can't lower a value of type " + probe.what + " yet"
			if !errors.As(err, &gap) || !strings.HasSuffix(err.Error(), want) {
				t.Fatalf("got %v, want NotYet ending %q", err, want)
			}
		})
	}
}

// Earlier representation stops can hide a mistaken read exemption. Ask the
// checker-backed read proof directly, without lowering a parameter first.
// Mutant run and restored: restore brandRead's private wide primitive set;
// the bigint, symbol and non-phantom string rows incorrectly become brands.
func TestBrandPrimitivesReadUsesPhantomProof(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct {
		name, receiver string
		brand          bool
	}{
		{"string", "string & { readonly __b: void }", true},
		{"number", "number & { readonly __b: void }", true},
		{"boolean", "boolean & { readonly __b: void }", true},
		{"optional receiver", "(string & { readonly __b: void }) | undefined | null", true},
		{"bigint", "bigint & { readonly __b?: undefined }", false},
		{"symbol", "symbol & { readonly __b?: undefined }", false},
		{"non-phantom string", "string & { readonly __b: number }", false},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			source := "function read(b: " + probe.receiver + "): void { b?.__b; }"
			program, err := load.Load([]string{writeAgreementSource(t, "main.ts", source)})
			if err != nil {
				t.Fatal(err)
			}
			file := program.Files()[0]
			checked, release := program.Checker(context.Background(), file)
			defer release()
			l := &lowering{program: program, checker: checked}
			read := file.Statements.Nodes[0].Body().AsBlock().Statements.Nodes[0].AsExpressionStatement().Expression
			if got := l.brandRead(read); got != probe.brand {
				t.Fatalf("brandRead(%s) = %v, want %v from phantomBase", probe.receiver, got, probe.brand)
			}
		})
	}
}

func TestBrandPrimitivesVoidMemberReadStops(t *testing.T) {
	t.Parallel()
	for _, primitive := range []string{"bigint", "symbol"} {
		t.Run(primitive, func(t *testing.T) {
			t.Parallel()
			source := "function read(b: " + primitive + " & { readonly __b: void }): void { console.log(String(b.__b)); }"
			_, err := lowerSourceAs(t, source, ".ts")
			var refused *Refused
			want := "main.ts:1:78: Adamic 0.1 refuses a read of a void value, as an argument; call it as a statement, or give the function a return type (adamic/no-void-read)"
			if !errors.As(err, &refused) || !strings.HasSuffix(err.Error(), want) {
				t.Fatalf("got %v, want a refusal ending %q", err, want)
			}
		})
	}
}
