package lower

import (
	"errors"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

func TestNumericEnumsAreOpen(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source string }{
		{"singleton", "enum E { A } function numeric(): number { return 42; } const e: E = numeric(); console.log(`${e}`);"},
		{"member_tags", "enum Kind { A, B } interface A { readonly kind: Kind.A; readonly value: string } interface B { readonly kind: Kind.B; readonly value: number } const a: A = { kind: Kind.A, value: 'text' }; function show(value: A | B): void { if(value.kind === Kind.B) { console.log(`${value.value + 1}`); } } show(a);"},
		{"coalesce", "enum Flags { None = 0, A = 1 << 0, B = 1 << 1, Pair = A | B } const map = new Map<string, Flags>(); console.log(`${map.get('high') ?? Flags.None}`);"},
		{"arithmetic", "enum E { A, B } function next(n: number): E { return n + 1; } let e: E = next(40); e++; e += 7; e <<= 1;"},
		{"signed_flags", "enum Flags { A = 1 << 31, B = 1 << 0 } const flags: Flags = ~Flags.A | Flags.B;"},
		{"cast", "enum SyntaxKind { First, Last } function kind(n: number): SyntaxKind { return n as SyntaxKind; }"},
		{"containers", "enum E { A, B } const values: E[] = [E.A]; const numbers: number[] = values; numbers.push(77); const map = new Map<string, E>(); const wide: Map<string, number> = map; wide.set('x', 99);"},
		{"optional_object", "enum E { A, B } function flags(box: { flags: E; other?: { flags: E } }): E { return box.other ? box.flags | box.other.flags : box.flags; }"},
	} {
		t.Log(probe.name)
		program, err := lowerSource(t, probe.source)
		if err != nil {
			t.Fatal(err)
		}
		name := "E"
		fields := []enumGuardField{{"A", float64(0)}, {"0", "A"}, {"B", float64(1)}, {"1", "B"}}
		switch probe.name {
		case "singleton":
			fields = fields[:2]
		case "member_tags":
			name = "Kind"
		case "coalesce":
			name = "Flags"
			fields = []enumGuardField{{"None", float64(0)}, {"0", "None"}, {"A", float64(1)}, {"1", "A"}, {"B", float64(2)}, {"2", "B"}, {"Pair", float64(3)}, {"3", "Pair"}}
		case "signed_flags":
			name = "Flags"
			fields = []enumGuardField{{"A", float64(-2147483648)}, {"-2147483648", "A"}, {"B", float64(1)}, {"1", "B"}}
		case "cast":
			name = "SyntaxKind"
			fields = []enumGuardField{{"First", float64(0)}, {"0", "First"}, {"Last", float64(1)}, {"1", "Last"}}
		}
		assertEnumGuardFields(t, program, name, fields)
		l, statements := enumGuardProof(t, probe.source)
		if !l.openNumericEnumType(l.checker.GetTypeAtLocation(statements[0].Name())) {
			t.Fatal("whole numeric enum must remain open")
		}
	}
}

func TestStringEnumsStayClosed(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"enum E { A = 'a', B = 'b' } function word(value: string): E { return value as E; }",
		"enum E { A = 'a', B = 'b' } const value: E = 'a' as E;",
	} {
		_, err := lowerSource(t, source)
		var refused *Refused
		if !errors.As(err, &refused) || (!strings.Contains(err.Error(), "enum-members") && !strings.Contains(err.Error(), "no-unchecked-cast")) {
			t.Fatalf("want closed string enum refusal, got %v", err)
		}
	}
}

func TestNumericEnumNeverProof(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct {
		name, source string
		guarded      bool
	}{
		{"member", "enum E { A, B } const member: E = E.A; switch(member) { case E.A: break; default: { const unreachable: never = member; console.log(`${unreachable}`); } }", false},
		{"parameter", "enum E { A, B } function show(e: E): number { switch(e) { case E.A: return 1; case E.B: return 2; default: { const unreachable: never = e; return 3; } } }", true},
		{"combination", "enum E { A = 1 << 0, B = 1 << 1 } const mask: E = E.A | E.B; switch(mask) { case E.A: break; case E.B: break; default: { const unreachable: never = mask; console.log(`${unreachable}`); } }", true},
		{"array_read", "enum E { A, B } function show(values: E[]): number { if(values[0] === undefined) return 0; if(values[0] === E.A) return 1; if(values[0] === E.B) return 2; const unreachable: never = values[0]; return 3; }", true},
		{"closed_return", "enum E { A, B } enum S { A = 'a' } function show(e: E): S { if (e === E.A) { return S.A; } if (e === E.B) { return S.A; } return e; }", true},
	} {
		t.Run(probe.name, func(t *testing.T) {
			program, err := lowerSource(t, probe.source)
			if err != nil {
				t.Fatal(err)
			}
			guarded := false
			for _, function := range program.Functions {
				if function.Name == "enum_never" {
					for _, statement := range function.Body {
						if _, ok := statement.(ir.Panic); ok {
							guarded = true
						}
					}
				}
			}
			if guarded != probe.guarded {
				t.Fatalf("guarded %v, want %v", guarded, probe.guarded)
			}
		})
	}
}

func TestNumericEnumLiteralPromises(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source string }{
		{"member_tag", "enum Kind { A, B } interface A { readonly kind: Kind.A; readonly value: string } interface B { readonly kind: Kind.B; readonly value: number } function numeric(): number { return 1; } const a: A = { kind: numeric(), value: 'text' }; function show(value: A | B): void { if (value.kind === Kind.B) { console.log(`${value.value + 1}`); } } show(a);"},
		{"narrowed_member", "enum E { A, B } function show(e: E): E.A { if(e === E.B) { return E.A; } return e; }"},
		{"plain_literal", "enum E { A } function numeric(): number { return 42; } const e: E = numeric(); const zero: 0 = e;"},
		{"indexed_subset", "enum E { A, B, C } function show(values: E[]): E.A | E.B { if(values[0] === undefined || values[0] === E.C) { return E.A; } return values[0]; }"},
		{"subset_member", "enum E { A, B, C } function show(e: E): E.A | E.B { if(e === E.C) { return E.A; } return e; }"},
		{"open_tag", "enum AKind { A } enum BKind { B = 1 } interface A { readonly kind: AKind; readonly value: string } interface B { readonly kind: BKind; readonly value: number } function numeric(): number { return 1; } const a: A = { kind: numeric(), value: 'text' }; function show(value: A | B): void { if (value.kind === BKind.B) { console.log(`${value.value + 1}`); } } show(a);"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			_, err := lowerSource(t, probe.source)
			var refused *Refused
			if !errors.As(err, &refused) {
				t.Fatalf("want literal-promise refusal, got %v", err)
			}
		})
	}
}
