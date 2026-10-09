package lower

import (
	"errors"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

func TestEmptyArrayInstantiations(t *testing.T) {
	t.Parallel()
	program, err := lowerSource(t, `function empty<T>(): T[] { return [] as T[]; }
interface Item { readonly text: string; }
empty<number>(); empty<Item>();`)
	if err != nil {
		t.Fatal(err)
	}
	elements := map[ir.Type]int{}
	for _, function := range program.Functions {
		if !strings.HasPrefix(function.Name, "empty_") {
			continue
		}
		returned := function.Body[0].(ir.Return)
		literal, ok := returned.Value.(ir.ArrayLiteral)
		if !ok || len(literal.Elements) != 0 {
			t.Fatalf("want empty literal, got %#v", returned.Value)
		}
		elements[literal.Element]++
	}
	if elements[ir.Number] != 1 || elements[ir.Object] != 1 || len(elements) != 2 {
		t.Fatalf("want number and object instantiations, got %v", elements)
	}
}

func TestEmptyArrayViewsStillRefuse(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source string }{
		{"mutable never alias", `const bottom: never[] = []; const wider: number[] = bottom;`},
		{"generic mutable result", `const bottom: never[] = []; function unsafe<T extends {}>(input: T[] | undefined): T[] { return input ?? bottom; } unsafe<number>(undefined);`},
		{"inferred mutable fallback", `const bottom: never[] = []; function unsafe<T extends {}>(input: T[] | undefined): readonly T[] { const result = input ?? bottom; return result; } unsafe<number>(undefined);`},
		{"nonempty generic cast", `function unsafe<T extends {}>(): T[] { return [{}] as T[]; } unsafe<{ readonly text: string }>();`},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSource(t, probe.source)
			var refused *Refused
			if !errors.As(err, &refused) {
				t.Fatalf("want refusal, got %v", err)
			}
		})
	}
}

func TestEmptyArrayScalarFallbacksUseTruthiness(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		`function fallback(value: number | undefined): number { return value || 7; } fallback(0);`,
		`function fallback(value: string | undefined): string { return value || 'missing'; } fallback('');`,
	} {
		program, err := lowerSource(t, source)
		if err != nil {
			t.Fatal(err)
		}
		for _, function := range program.Functions {
			if strings.HasPrefix(function.Name, "fallback") {
				returned := function.Body[0].(ir.Return)
				if _, wrong := returned.Value.(ir.Coalesce); wrong {
					t.Fatal("scalar OR lost zero/empty-string truthiness")
				}
			}
		}
	}
}
