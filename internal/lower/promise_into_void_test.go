package lower

import (
	"errors"
	"strings"
	"testing"
)

// A promise that reaches a discard by a road :355's call check doesn't see is :355's unawaited task one
// step removed, and a rejection exits 0 natively where Node exits 1 (#8j8ecw0, Runtime Performance's
// probe on #995spev, Language's roads on #8j8ecw0). Each road is refused in .a and .ts at the value's
// line: a promise-returning function seen as a void-returning one, a promise seen as a type that names
// no promise (Language's (a)), and an expression thrown away whose type holds a promise inside (her
// (b)). A promise kept where its type says so, an async return adopting one, Promise.all over a map, a
// void function into a void slot and a number-returning function into a void slot (:325's discard of a
// plain value) still lower. Mutants: either rule's registration removed, and its rows lower again.
func TestPromiseIntoVoidIsRefused(t *testing.T) {
	t.Parallel()
	fails := "async function fails(): Promise<void> {\n\tawait Promise.resolve();\n\tthrow new Error('rejected');\n}\n"
	for _, probe := range []struct {
		name, source string
		line         string // the refused value's line, or "" where the program lowers
	}{
		{"declaration", "const run: () => void = fails;\nrun();\n", ":5:"},
		{"argument", "function apply(f: () => void): void {\n\tf();\n}\napply(fails);\n", ":8:"},
		{"library callback", "[1].forEach(async (): Promise<void> => {\n\tawait fails();\n});\n", ":5:"},
		{"field", "const holder: { run: () => void } = { run: fails };\nholder.run();\n", ":5:"},
		{"element", "const runs: (() => void)[] = [fails];\nruns[0]();\n", ":5:"},
		{"return", "function view(): () => void {\n\treturn fails;\n}\nview()();\n", ":6:"},
		{"assertion", "const run = fails as () => void;\nrun();\n", ":5:"},
		{"union member", "const run: (() => void) | number = fails;\nif (typeof run === 'function') {\n\trun();\n}\n", ":5:"},
		{"parameter", "function each(f: () => void): void {\n\tf();\n}\nconst take: (f: () => Promise<void>) => void = each;\ntake(fails);\n", ":8:"},
		{"unknown view", "const run: () => unknown = fails;\nrun();\n", ":5:"},
		{"void or undefined view", "const run: () => void | undefined = fails;\nrun();\n", ":5:"},
		{"unknown value", "const held: unknown = fails();\nconsole.log(typeof held);\n", ":5:"},
		{"discarded map", "[0].map(fails);\n", ":5:"},
		{"discarded and", "const ok: boolean = true;\nok && fails();\n", ":6:"},
		{"discarded comma", "const n: number = (fails(), 0);\nconsole.log(String(n));\n", ":5:"},
		{"printed", "console.log(fails());\n", ":5:"},
		{"through an identity", "function identity<T>(value: T): T {\n\treturn value;\n}\nconst kept = identity(fails());\nawait kept.catch((): void => {});\n", ""},
		{"async return adopts", "async function later(): Promise<void> {\n\treturn fails();\n}\nawait later().catch((): void => {});\n", ""},
		{"awaited all", "await Promise.all([0].map(fails)).catch((): void => {});\n", ""},
		{"kept in a const", "const kept: Promise<void> = fails();\nawait kept.catch((): void => {});\n", ""},
		{"kept as a promise", "const run: () => Promise<void> = fails;\nawait run().catch((): void => {});\n", ""},
		{"void into void", "function quiet(): void {}\nconst run: () => void = quiet;\nrun();\n", ""},
		{"awaited directly", "await fails().catch((): void => {});\n", ""},
		{"a value discarded", "function count(): number {\n\treturn 1;\n}\nconst run: () => void = count;\nrun();\n", ""},
	} {
		for _, extension := range []string{".a", ".ts"} {
			t.Run(probe.name+extension, func(t *testing.T) {
				t.Parallel()
				_, err := lowerSourceAs(t, fails+probe.source, extension)
				if probe.line == "" {
					var refused *Refused
					if errors.As(err, &refused) && strings.Contains(refused.What, "unawaited async task") {
						t.Fatalf("want no promise-into-void refusal, got %v", err)
					}
					return
				}
				var refused *Refused
				if !errors.As(err, &refused) || !strings.HasPrefix(refused.What, "an unawaited async task: a ") || !strings.Contains(refused.Where, probe.line) {
					t.Fatalf("want the promise-into-void refusal at line %s, got %v", probe.line, err)
				}
			})
		}
	}
}
