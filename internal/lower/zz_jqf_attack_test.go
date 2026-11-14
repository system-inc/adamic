package lower

import "testing"

// A probe, never for landing: Optimizer's read of #jqf1drp at 52aafdc6, the closure under generics.
const jqfBase = `interface Circle {
    readonly kind: "circle";
    readonly radius: number;
}
interface Square {
    readonly kind: "square";
    readonly side: number;
}
type Shape = Circle | Square;
function find<T, U extends T>(array: readonly T[] | undefined, predicate: (element: T, index: number) => element is U, startIndex?: number): U | undefined;
function find<T>(array: readonly T[] | undefined, predicate: (element: T, index: number) => boolean, startIndex?: number): T | undefined;
function find<T>(array: readonly T[] | undefined, predicate: (element: T, index: number) => boolean, startIndex?: number): T | undefined {
    if (array === undefined) return undefined;
    for (let i = startIndex ?? 0; i < array.length; i++) {
        const value = array[i]!;
        if (predicate(value, i)) {
            return value;
        }
    }
    return undefined;
}
function isCircle(shape: Shape): shape is Circle {
    return shape.kind === "circle";
}
function isSquare(shape: Shape): shape is Square {
    return shape.kind === "square";
}
const shapes: Shape[] = [{ kind: "square", side: 2 }, { kind: "circle", radius: 3 }];
function getFirst<T extends Shape>(xs: readonly Shape[], predicate: (shape: Shape) => shape is T): T | undefined {
    return find(xs, predicate);
}
`

func TestZZJqfAttack(t *testing.T) {
	for _, probe := range []struct{ name, source, want string }{
		// T stands only in a function-type parameter, so its instances may share one body.
		{"unread T, two instances", "function has<T extends Shape>(xs: readonly Shape[], test: (shape: Shape) => shape is T): boolean {\n    return find(xs, test) !== undefined;\n}\nconsole.log(String(has(shapes, isCircle)));\nconsole.log(String(has(shapes, isSquare)));\n", "true\ntrue\n"},
		{"read T, two instances", "const c = getFirst(shapes, isCircle);\nconsole.log(c === undefined ? \"none\" : `radius ${String(c.radius)}`);\nconst s = getFirst(shapes, isSquare);\nconsole.log(s === undefined ? \"none\" : `side ${String(s.side)}`);\n", "radius 3\nside 2\n"},
		{"captured parameter", "function outer(test: (shape: Shape) => shape is Circle): string {\n    const run = (): Circle | undefined => find(shapes, test);\n    const c = run();\n    return c === undefined ? \"none\" : `radius ${String(c.radius)}`;\n}\nconsole.log(outer(isCircle));\n", "radius 3\n"},
		{"forwarded twice, generic", "function getFirst2<T extends Shape>(xs: readonly Shape[], p: (shape: Shape) => shape is T): T | undefined {\n    return getFirst(xs, p);\n}\nconst c = getFirst2(shapes, isCircle);\nconsole.log(c === undefined ? \"none\" : `radius ${String(c.radius)}`);\nconst s = getFirst2(shapes, isSquare);\nconsole.log(s === undefined ? \"none\" : `side ${String(s.side)}`);\n", "radius 3\nside 2\n"},
		{"unread T through a generic find of T", "function count<T extends Shape>(xs: readonly Shape[], test: (shape: Shape) => shape is T): number {\n    let n = 0;\n    for (const x of xs) {\n        if (find([x], test) !== undefined) {\n            n = n + 1;\n        }\n    }\n    return n;\n}\nconsole.log(String(count(shapes, isCircle)));\nconsole.log(String(count(shapes, isSquare)));\n", "1\n1\n"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			program := lowersAndAgreesWithNodeAs(t, jqfBase+probe.source, ".ts")
			code, out, errs := brandedNative(t, program)
			t.Logf("PROBE %s native: exit %d stdout %q stderr %q", probe.name, code, out, errs)
			if code != 0 || out != probe.want || errs != "" {
				t.Errorf("PROBE %s: want %q", probe.name, probe.want)
			}
		})
	}
}
