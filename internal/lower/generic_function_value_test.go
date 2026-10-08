package lower

import (
	"strings"
	"testing"
)

func TestGenericFunctionValueBoundaries(t *testing.T) {
	t.Parallel()
	for _, test := range []struct{ name, source, reason string }{
		{"polymorphic", `function id<T>(x: T): T { return x; } const f = id;`, "concrete contextual signature"},
		{"phantom", `function id<T, U>(x: T): T { return x; } const f: (x: number) => number = id;`, "unresolved type parameter"},
		{"arity", `function id<T>(x: T, extra: number = 1): T { return x; } const f: (x: number) => number = id;`, "parameter count differs"},
		{"parameter", `function id<T>(x: T, y: { readonly n: number }): T { return x; } const f: (x: number, y: { readonly n: number; readonly extra: string }) => number = id;`, "instantiated parameters differ"},
		{"result", `function id<T>(x: T): T { return x; } const f: (x: number) => number | undefined = id;`, "instantiated result differs"},
		{"identity", `function id<T>(x: T): T { return x; } const f: (x: number) => number = id; const g: (x: string) => string = id; console.log(g('x'));`, "one shared source identity"},
	} {
		t.Run(test.name, func(t *testing.T) {
			_, err := lowerSource(t, test.source)
			if err == nil || !strings.Contains(err.Error(), test.reason) {
				t.Fatalf("want %s, got %v", test.reason, err)
			}
		})
	}
}
