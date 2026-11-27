package lower

import (
	"errors"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

// A member cast whose source holds undefined beside a represented union (#b5w3ycg slice 2, absentArm) lowers
// in .ts to the absent check, then the tag check: one CheckedCast whose value is the Coalesce that panics with
// the cast's own message.
func TestAbsentArmCastsLower(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source, message string }{
		{"member", "enum K { A = 1, B = 2 } type A = { readonly kind: K.A; readonly n: number }; type B = { readonly kind: K.B; readonly n: number }; function f(x: A | B | undefined): A { return x as A; }", "cast failed: this A | B | undefined is not a A"},
		{"sub-union", "enum K { A = 1, B = 2, C = 3 } type A = { readonly kind: K.A }; type B = { readonly kind: K.B }; type C = { readonly kind: K.C }; function f(x: A | B | C | undefined): A | C { return x as A | C; }", "cast failed: this A | B | C | undefined is not a A | C"},
		{"string tags", "type A = { readonly kind: 'a'; readonly n: number }; type B = { readonly kind: 'b'; readonly s: string }; function f(x: A | B | undefined): B { return x as B; }", "cast failed: this A | B | undefined is not a B"},
		{"optional parameter", "enum K { A = 1, B = 2 } type A = { readonly kind: K.A }; type B = { readonly kind: K.B }; function f(x?: A | B): A { return x as A; }", "cast failed: this A | B | undefined is not a A"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			program, err := lowerSourceAs(t, probe.source, ".ts")
			if err != nil {
				t.Fatal(err)
			}
			casts := 0
			for _, function := range program.Functions {
				walk(function.Body, func(node any) bool {
					cast, ok := node.(ir.CheckedCast)
					if !ok {
						return true
					}
					casts++
					absent, ok := cast.Value.(ir.Coalesce)
					if !ok || absent.Panic == nil || absent.Fallback != nil {
						t.Fatalf("the cast's value is %T, want the Coalesce that panics before the tag is read", cast.Value)
					}
					if text := program.Strings[absent.Panic.(ir.StringConstant).Index]; text != probe.message || cast.Message != probe.message {
						t.Fatalf("absent message %q, tag message %q, want both %q", text, cast.Message, probe.message)
					}
					return true
				})
			}
			if casts != 1 {
				t.Fatalf("got %d checked casts, want 1", casts)
			}
		})
	}
}

// What absentArm leaves as it was: a null beside the members, a writable tag, and a tag the checker's duplicates
// can't tell apart. A single member beside undefined, cast to that member, is presentCast's road
// (cast_drops_nullish_test.go), and a target that keeps undefined is keepsNullishCast's (cast_keeps_nullish_test.go).
// Each row pins its own cause's stop, as castRoad's untagged rows do: a null beside the members has no literal tag,
// so the untagged arm's NotYet stops it first (cast_untagged_union.go, #axqgdck), and the other two keep the
// unchecked-cast refusal. These rows never ran until the file's name stopped reading as `GOARCH=arm` (#b5w3ycg).
func TestAbsentArmCastsStayRefused(t *testing.T) {
	t.Parallel()
	members := "enum K { A = 1, B = 2 } type A = { readonly kind: K.A }; type B = { readonly kind: K.B }; "
	untagged, unchecked := "no literal tag tells apart", "adamic/no-unchecked-cast"
	for _, probe := range []struct {
		name, source, extension, want string
		refused                       bool
	}{
		{"null beside", members + "function f(x: A | B | null): A { return x as A; }", ".ts", untagged, false},
		{"null and undefined beside", members + "function f(x: A | B | null | undefined): A { return x as A; }", ".ts", untagged, false},
		{"writable tag", "enum K { A = 1, B = 2 } type A = { kind: K.A }; type B = { kind: K.B }; function f(x: A | B | undefined): A { return x as A; }", ".ts", unchecked, true},
		{"duplicate tags", "enum K { A = 1, B = 1 } type A = { readonly kind: K.A; readonly n: number }; type B = { readonly kind: K.B; readonly s: string }; function f(x: A | B | undefined): A { return x as A; }", ".ts", unchecked, true},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSourceAs(t, probe.source, probe.extension)
			var refused *Refused
			var notYet *NotYet
			switch {
			case probe.refused && (!errors.As(err, &refused) || !strings.Contains(err.Error(), probe.want)):
				t.Fatalf("want a refusal holding %q, got %v", probe.want, err)
			case !probe.refused && (!errors.As(err, &notYet) || !strings.Contains(err.Error(), probe.want)):
				t.Fatalf("want NotYet holding %q, got %v", probe.want, err)
			}
		})
	}
}

// In .a the same cast is a member cast, refused as #m8dvvmf refuses every one, and the narrowing its fix offers
// reads the tag through ?., so that it type-checks on an operand that may be undefined and fails it too.
func TestAbsentArmCastInDotAIsRefusedWithItsNarrowing(t *testing.T) {
	t.Parallel()
	_, err := lowerSourceAs(t, "enum K { A = 1, B = 2 } type A = { readonly kind: K.A }; type B = { readonly kind: K.B }; function f(x: A | B | undefined): A { return x as A; }", ".a")
	var refused *Refused
	if !errors.As(err, &refused) || !strings.Contains(err.Error(), "in .a whose checks no proof erases") || !strings.Contains(err.Error(), "if (x?.kind !== K.A) panic('why')") {
		t.Fatalf("want .a's member-cast refusal offering x?.kind, got %v", err)
	}
}

// The absent field's premise (Language's v03 on #ha4hbqt): a tag read through a getter stops at the checked view
// before either test runs, with and without an undefined beside the members, so the absent test and the tag test
// can't disagree through one. If that rule lifts, this goes red and the premise needs reading again.
func TestAbsentArmGetterTagStopsFirst(t *testing.T) {
	t.Parallel()
	for _, source := range []string{"A | B | undefined", "A | B"} {
		t.Run(source, func(t *testing.T) {
			t.Parallel()
			program := "enum K { A = 1, B = 2 } interface A { readonly kind: K.A; readonly n: number } interface B { readonly kind: K.B; readonly n: number } class Lying { get kind(): K.A { return K.A; } readonly n = 5; } function f(x: " + source + "): A { return x as A; } console.log(String(f(new Lying()).n));"
			_, err := lowerSourceAs(t, program, ".ts")
			if err == nil || !strings.Contains(err.Error(), "a getter in a checked field contract") {
				t.Fatalf("want the stop at a getter in a checked field contract, got %v", err)
			}
		})
	}
}

// absentObjectModel is tsc's `node as SourceFile` of a `Node | undefined` (cause K of #eg3krrp): one object beside
// undefined, whose kind is the open SyntaxKind, cast to an interface that narrows kind to one member of it.
const absentObjectModel = "enum SyntaxKind { Identifier, SourceFile, Block } interface Node { readonly kind: SyntaxKind; readonly pos: number } interface SourceFile extends Node { readonly kind: SyntaxKind.SourceFile; readonly fileName: string } function f(x: Node | undefined): string { return (x as SourceFile).fileName; } "

// A downcast of the one object beside undefined lowers in .ts to the absent check, then that object's own interface
// downcast by its kind tag: an honest SourceFile reads through it and agrees with Node, on the sanitized native too.
// Mutant: absentArm wanting two or more present members again, and the cast stops with the untagged union's NotYet.
func TestAbsentArmObjectDowncastLowers(t *testing.T) {
	t.Parallel()
	program := lowersAndAgreesWithNodeAs(t, absentObjectModel+"const file: SourceFile = { kind: SyntaxKind.SourceFile, pos: 0, fileName: 'a.ts' }; console.log(f(file));", ".ts")
	if code, stdout, stderr := brandedNative(t, program); code != 0 || stdout != "a.ts\n" {
		t.Fatalf("sanitized native exit %d, stdout %q, stderr %.300q; want a.ts, as Node prints", code, stdout, stderr)
	}
}

// undefined fails the cast before its tag is read, with the cast's own message, which names the source as written.
// Mutant: castValue's absent check dropped for this road, and native reads the tag of the undefined it was handed.
func TestAbsentArmObjectDowncastStopsUndefined(t *testing.T) {
	t.Parallel()
	program, err := lowerSourceAs(t, absentObjectModel+"console.log(f(undefined));", ".ts")
	if err != nil {
		t.Fatal(err)
	}
	if code, stdout, stderr := brandedNative(t, program); code == 0 || stdout != "" || !strings.Contains(stderr, "cast failed: this Node | undefined is not a SourceFile") {
		t.Fatalf("sanitized native exit %d, stdout %q, stderr %.300q; want a stop at the cast's absent check, in row 9's words", code, stdout, stderr)
	}
}

// A Node of another kind that holds the fileName the view reads fails at the cast's tag test, which interfaceCast
// words from the present object, as a cast that keeps nullish does. Mutant: the present road skipped, so the value
// passes the absent check unviewed, and native reads the Block's fileName and prints b.ts.
func TestAbsentArmObjectDowncastStopsAnotherKind(t *testing.T) {
	t.Parallel()
	program, err := lowerSourceAs(t, absentObjectModel+"interface Block extends Node { readonly kind: SyntaxKind.Block; readonly fileName: string } const block: Block = { kind: SyntaxKind.Block, pos: 0, fileName: 'b.ts' }; console.log(f(block));", ".ts")
	if err != nil {
		t.Fatal(err)
	}
	if code, stdout, stderr := brandedNative(t, program); code == 0 || stdout != "" || !strings.Contains(stderr, "cast failed: this Node is not a SourceFile") {
		t.Fatalf("sanitized native exit %d, stdout %q, stderr %.300q; want a stop at the cast's tag test, in row 9's words", code, stdout, stderr)
	}
}

// A null beside the one object is outside absentArm, so the cast keeps the untagged union's NotYet it had, and never
// a refusal. Mutant: absentArm taking null as it takes undefined, and the cast lowers.
func TestAbsentArmObjectDowncastBesideNullStops(t *testing.T) {
	t.Parallel()
	_, err := lowerSourceAs(t, strings.Replace(absentObjectModel, "x: Node | undefined", "x: Node | null", 1), ".ts")
	const want = "a checked view of SourceFile over the union Node | null, whose members no literal tag tells apart"
	var refused *Refused
	var notYet *NotYet
	if errors.As(err, &refused) || !errors.As(err, &notYet) || !strings.Contains(err.Error(), want) {
		t.Fatalf("want NotYet holding %q, never a refusal, got %v", want, err)
	}
}
