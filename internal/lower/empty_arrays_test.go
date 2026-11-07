package lower

import (
	"errors"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

func TestEmptyArrayInstantiations(test *testing.T) {
	test.Parallel()
	program, lowerSourceError := lowerSource(test, `function empty<T>(): T[] { return [] as T[]; }
interface Item { readonly text: string; }
empty<number>(); empty<Item>();`)
	if lowerSourceError != nil {
		test.Fatal(lowerSourceError)
	}
	elements := map[ir.Type]int{}
	for _, function := range program.Functions {
		if !strings.HasPrefix(function.Name, "empty_") {
			continue
		}
		returned := function.Body[0].(ir.Return)
		literal, ok := returned.Value.(ir.ArrayLiteral)
		if !ok || len(literal.Elements) != 0 {
			test.Fatalf("want empty literal, got %#v", returned.Value)
		}
		elements[literal.Element]++
	}
	if elements[ir.Number] != 1 || elements[ir.Object] != 1 || len(elements) != 2 {
		test.Fatalf("want number and object instantiations, got %v", elements)
	}
}

func TestEmptyArrayViewsStillRefuse(test *testing.T) {
	test.Parallel()
	for _, probe := range []struct{ name, source string }{
		{"mutable never alias", `const bottom: never[] = []; const wider: number[] = bottom;`},
		{
			"generic mutable result",
			`const bottom: never[] = []; function unsafe<T extends {}>(input: T[] | undefined): T[] { return input ?? bottom; } unsafe<number>(undefined);`,
		},
		{
			"inferred mutable fallback",
			`const bottom: never[] = []; function unsafe<T extends {}>(input: T[] | undefined): readonly T[] { const result = input ?? bottom; return result; } unsafe<number>(undefined);`,
		},
		{
			"nonempty generic cast",
			`function unsafe<T extends {}>(): T[] { return [{}] as T[]; } unsafe<{ readonly text: string }>();`,
		},
	} {
		test.Run(probe.name, func(test *testing.T) {
			test.Parallel()
			_, lowerSourceError := lowerSource(test, probe.source)
			var refused *Refused
			if !errors.As(lowerSourceError, &refused) {
				test.Fatalf("want refusal, got %v", lowerSourceError)
			}
		})
	}
}

func TestEmptyArrayScalarFallbacksUseTruthiness(test *testing.T) {
	test.Parallel()
	for _, source := range []string{
		`function fallback(value: number | undefined): number { return value || 7; } fallback(0);`,
		`function fallback(value: string | undefined): string { return value || 'missing'; } fallback('');`,
	} {
		program, lowerSourceError := lowerSource(test, source)
		if lowerSourceError != nil {
			test.Fatal(lowerSourceError)
		}
		for _, function := range program.Functions {
			if strings.HasPrefix(function.Name, "fallback") {
				returned := function.Body[0].(ir.Return)
				if _, wrong := returned.Value.(ir.Coalesce); wrong {
					test.Fatal("scalar OR lost zero/empty-string truthiness")
				}
			}
		}
	}
}
