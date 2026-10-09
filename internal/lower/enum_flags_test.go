package lower

import (
	"errors"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

const flagDeclaration = "enum Flags { None = 0, A = 1 << 0, B = 1 << 1, High = 1 << 30, Both = A | B } "

func TestFlagEnumsOpen(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source, reason string }{
		{"complement", flagDeclaration + "const value: Flags = ~Flags.A;", "enum-flags"},
		{"arithmetic", flagDeclaration + "const value: Flags = Flags.A + 1;", "enum-flags"},
		{"increment", flagDeclaration + "let flags: Flags = Flags.A; flags++;", "enum-flags"},
		{"shift_update", flagDeclaration + "let flags: Flags = Flags.A; flags <<= 1;", "enum-flags"},
		{"shift_result", flagDeclaration + "const flags: Flags = Flags.A << 1;", "enum-flags"},
		{"sign_bit", "enum Flags { A = 1 << 31 }", "non-negative int32 bound"},
		{"implicit", "enum Color { Red, Green, Blue } const color: Color = Color.Red | Color.Green;", "closed union"},
		{"numeric_shape", "enum Color { Red = 0, Green = 1, Blue = 2 } const color: Color = Color.Green | Color.Blue;", "closed union"},
		{"arithmetic_shape", "enum Color { Red = 0, Green = 1 << 0, Blue = (1 << 0) + 1 } const color: Color = Color.Green | Color.Blue;", "closed union"},
		{"switch_default", flagDeclaration + "function show(flags: Flags): void { switch (flags) { case Flags.None: break; case Flags.A: break; case Flags.B: break; case Flags.High: break; case Flags.Both: break; } }", "flag-enum switch without a default"},
		{"cross_enum_left", flagDeclaration + "enum Other { A = 1 << 0, B = 1 << 1 } const flags: Flags = Flags.A | Other.B;", "enum-flags"},
		{"cross_enum_right", flagDeclaration + "enum Other { A = 1 << 0, B = 1 << 1 } const flags: Other = Flags.A | Other.B;", "enum-flags"},
		{"or_number", flagDeclaration + "function open(n: number): Flags { return Flags.A | n; }", "enum-flags"},
		{"xor_number", flagDeclaration + "function open(n: number): Flags { return n ^ Flags.A; }", "enum-flags"},
		{"and_numbers", flagDeclaration + "function open(n: number): Flags { return n & 3; }", "enum-flags"},
		{"narrowed_member", "enum Flags { A = 1 << 0, B = 1 << 1 } function wrong(flags: Flags): Flags.B { if (flags === Flags.A) { return Flags.B; } return flags; }", "enum-flags"},
		{"compound_or_number", flagDeclaration + "let flags: Flags = Flags.A; flags |= 8;", "enum-flags"},
		{"compound_xor_number", flagDeclaration + "let flags: Flags = Flags.A; flags ^= 8;", "enum-flags"},
		{"mutable_alias", flagDeclaration + "let combined = Flags.A | Flags.B; combined = 99; const flags: Flags = combined;", "enum-flags"},
		{"bare_number", flagDeclaration + "const flags: Flags = 0;", "enum-flags"},
		{"member_slot", flagDeclaration + "const flags: Flags.A = Flags.A | Flags.B;", "enum-flags"},
		{"array_view", flagDeclaration + "const flags: Flags[] = [Flags.A]; const numbers: number[] = flags;", "invariant-mutable"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSource(t, probe.source)
			if probe.name == "narrowed_member" || probe.name == "member_slot" {
				var refused *Refused
				if !errors.As(err, &refused) {
					t.Fatalf("want literal promise refusal, got %v", err)
				}
			} else if err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestFlagEnumsDomain(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source string }{
		{"bitwise_updates", "let flags: Flags = Flags.A; flags |= Flags.B; flags ^= Flags.High; flags &= ~Flags.A;"},
		{"alias", "const combined = Flags.A | Flags.B; const flags: Flags = combined;"},
		{"or", "function combine(a: Flags, b: Flags): Flags { return a | b; }"},
		{"xor", "function toggle(a: Flags, b: Flags): Flags { return a ^ b; }"},
		{"and_left", "function mask(a: Flags, n: number): Flags { return a & n; }"},
		{"and_right", "function mask(a: Flags, n: number): Flags { return n & a; }"},
		{"and_complement", "function mask(a: Flags): Flags { return a & ~Flags.A; }"},
		{"nested", "function mask(a: Flags, b: Flags): Flags { return (a | b) & ~Flags.A; }"},
		{"field", "const box: { value: Flags } = { value: Flags.A | Flags.B }; box.value = box.value ^ Flags.High;"},
		{"array", "const flags: Flags[] = [Flags.A | Flags.B]; flags.push(Flags.A ^ Flags.High);"},
		{"map", "const flags = new Map<string, Flags>([['x', Flags.A | Flags.B]]); flags.set('y', Flags.High | Flags.B);"},
		{"parameter", "function use(flags: Flags): void {} use(Flags.A | Flags.B);"},
		{"optional", "function use(flags: Flags | undefined): void {} use(Flags.A | Flags.B);"},
		{"switch", "function use(flags: Flags): void { switch(flags) { case Flags.A: break; default: break; } }"},
	} {
		t.Log(probe.name)
		program, err := lowerSource(t, flagDeclaration+probe.source)
		if err != nil {
			t.Fatal(err)
		}
		assertEnumGuardFields(t, program, "Flags", flagGuardFields)
	}
	assertEnumGuardDomains(t)
}

func TestEnumNeverDefault(t *testing.T) {
	t.Parallel()
	program, err := lowerSource(t, "enum Color { Red, Green, Blue } function describe(value: Color): string { switch(value) { case Color.Red: return 'r'; case Color.Green: return 'g'; case Color.Blue: return 'b'; default: { const unreachable: never = value; return unreachable; } } }")
	if err != nil {
		t.Fatal(err)
	}
	assertEnumGuardFields(t, program, "Color", []enumGuardField{{"Red", float64(0)}, {"0", "Red"}, {"Green", float64(1)}, {"1", "Green"}, {"Blue", float64(2)}, {"2", "Blue"}})
	for _, function := range program.Functions {
		if function.Name == "enum_never" {
			for _, statement := range function.Body {
				if _, ok := statement.(ir.Panic); ok {
					return
				}
			}
		}
	}
	t.Fatal("open numeric enum default lost its runtime never guard")
}

func TestFlagEnumLiteralSpellings(t *testing.T) {
	t.Parallel()
	source := "enum Flags { None = 0x0, A = 0x1 << 0b0, High = 0o1 << 0x1e } const flags: Flags = Flags.A | Flags.High;"
	program, err := lowerSource(t, source)
	if err != nil {
		t.Fatal(err)
	}
	assertEnumGuardFields(t, program, "Flags", []enumGuardField{{"None", float64(0)}, {"0", "None"}, {"A", float64(1)}, {"1", "A"}, {"High", float64(1073741824)}, {"1073741824", "High"}})
	assertEnumGuardFlag(t, source)
}

func TestFlagEnumMemberAliases(t *testing.T) {
	t.Parallel()
	source := "enum Flags { None = 0, A = 1 << 0, B = 1 << 1, Alias = A, Qualified = Flags.B } const flags: Flags = Flags.Alias | Flags.Qualified;"
	program, err := lowerSource(t, source)
	if err != nil {
		t.Fatal(err)
	}
	assertEnumGuardFields(t, program, "Flags", []enumGuardField{{"None", float64(0)}, {"0", "None"}, {"A", float64(1)}, {"1", "Alias"}, {"B", float64(2)}, {"2", "Qualified"}, {"Alias", float64(1)}, {"Qualified", float64(2)}})
	assertEnumGuardFlag(t, source)
}

func TestEnumNameEnumeration(t *testing.T) {
	t.Parallel()
	source := "enum Names { A, B, Alias = B } const names = Names; for (const name in names) { console.log(name); }"
	program, err := lowerSource(t, source)
	if err != nil {
		t.Fatal(err)
	}
	assertEnumGuardFields(t, program, "Names", []enumGuardField{{"A", float64(0)}, {"0", "A"}, {"B", float64(1)}, {"1", "Alias"}, {"Alias", float64(1)}})
	loop := enumGuardLoop(t, program)
	if _, ok := loop.Iterable.(ir.ObjectKeys); !ok || loop.Element != ir.String {
		t.Fatalf("enum name iteration = %#v, want string iteration over ObjectKeys", loop)
	}
	enumGuardNodeOutput(t, source, program, "0\n1\nA\nB\nAlias\n", true)
}

func TestFlagEnumInlineIteration(t *testing.T) {
	t.Parallel()
	source := flagDeclaration + "function use(box: { flags: Flags }): void { console.log(`${box.flags}`); } for (const flags of [Flags.None, Flags.A | Flags.B]) { use({flags}); }"
	program, err := lowerSource(t, source)
	if err != nil {
		t.Fatal(err)
	}
	assertEnumGuardFields(t, program, "Flags", flagGuardFields)
	loop := enumGuardLoop(t, program)
	array, ok := loop.Iterable.(ir.ArrayLiteral)
	if !ok || len(array.Elements) != 2 || array.Element != ir.Number {
		t.Fatalf("inline enum iterable = %#v, want two numeric elements", loop.Iterable)
	}
	first, ok := array.Elements[0].(ir.Property)
	if !ok || first.Name != "None" {
		t.Fatalf("first inline element = %#v, want Flags.None", array.Elements[0])
	}
	second, ok := array.Elements[1].(ir.Binary)
	if !ok || second.Operator != ir.BitOr {
		t.Fatalf("second inline element = %#v, want bitwise OR", array.Elements[1])
	}
	left, leftOK := second.Left.(ir.Property)
	right, rightOK := second.Right.(ir.Property)
	if !leftOK || !rightOK || left.Name != "A" || right.Name != "B" {
		t.Fatalf("second inline operands = %#v, %#v, want Flags.A and Flags.B", second.Left, second.Right)
	}
	enumGuardNodeOutput(t, source, program, "0\n3\n", true)
}

func TestFlagEnumAliasBoundaries(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"enum Other { A = 1 << 0 } enum Flags { None = 0, A = 1 << 0, Alias = Other.A } const flags: Flags = Flags.A | Flags.Alias;",
		flagDeclaration + "for (const flags of [Flags.A, 99]) { const value: Flags = flags; }",
		flagDeclaration + "const numbers: number[] = [Flags.A]; numbers.push(99); for (const flags of numbers) { const value: Flags = flags; }",
		flagDeclaration + "for (let flags of [Flags.A | Flags.B]) { flags = 99; const value: Flags = flags; }",
		flagDeclaration + "const combined = Flags.A | Flags.B; function use(box: { flags: Flags.A }): void {} use({flags: combined});",
		flagDeclaration + "interface Node { flags: Flags } type Mutable<T> = { -readonly [K in keyof T]: T[K] }; function set<T extends Node>(node: T, flags: Flags): void { (node as Mutable<T>).flags = flags; }",
	} {
		_, err := lowerSource(t, source)
		if strings.Contains(source, "Mutable<T>") || strings.Contains(source, "flags: Flags.A") {
			var refused *Refused
			if !errors.As(err, &refused) {
				t.Fatalf("want Refused, got %v for %s", err, source)
			}
		} else if err != nil {
			t.Fatal(err)
		}
	}
}
