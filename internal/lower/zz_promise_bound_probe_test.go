package lower

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/load"
)

// A probe for #96pc420, never landed: which relation shape does refusePromiseIntoVoid alone walk deep,
// before refuseVoidWidening's and refuseAnyRelation's own bounds of 64 stop it? Each shape is lowered as
// .ts at four depths and one line is printed per lowering. It never fails: a load error is printed too.
func TestZZPromiseBoundProbe(t *testing.T) {
	t.Parallel()
	shapes := []struct {
		name string
		body string
	}{
		{"initializer", "const a: A0 = makeA();\nconst b: B0 = a;\nconsole.log(String(b.v));\n"},
		{"async-return", "async function load(): Promise<A0> {\n\treturn makeA();\n}\nasync function view(): Promise<B0> {\n\treturn load();\n}\nconst b: B0 = await view();\nconsole.log(String(b.v));\n"},
		{"async-return-awaited", "async function load(): Promise<A0> {\n\treturn makeA();\n}\nasync function view(): Promise<B0> {\n\treturn await load();\n}\nconst b: B0 = await view();\nconsole.log(String(b.v));\n"},
		{"argument", "function take(run: (b: B0) => void): void {\n\trun(makeB());\n}\ntake((a: A0): void => {\n\tconsole.log(String(a.v));\n});\n"},
		{"promise-field", "async function load(): Promise<A0> {\n\treturn makeA();\n}\nconst held: { p: Promise<A0> } = { p: load() };\nconst seen: { p: Promise<B0> } = held;\nconst b: B0 = await seen.p;\nconsole.log(String(b.v));\n"},
	}
	for _, shape := range shapes {
		for _, depth := range []int{70, 100, 200, 300} {
			t.Run(shape.name+"/"+strconv.Itoa(depth), func(t *testing.T) {
				t.Parallel()
				source := probeNestedTypes("A", depth) + probeNestedTypes("B", depth) +
					"function makeA(): A0 {\n\tthrow new Error('made');\n}\nfunction makeB(): B0 {\n\tthrow new Error('made');\n}\n" + shape.body
				fmt.Printf("PROBE %s %d: %s\n", shape.name, depth, probeLower(t, source))
			})
		}
	}
	t.Run("self-referential", func(t *testing.T) {
		t.Parallel()
		source := "interface ListA {\n\tnext: ListA | undefined;\n\trun: () => void;\n}\ninterface ListB {\n\tnext: ListB | undefined;\n\trun: () => void;\n}\n" +
			"const a: ListA = { next: undefined, run: (): void => {} };\nconst b: ListB = a;\nb.run();\nconsole.log(String(b.next === undefined));\n"
		fmt.Printf("PROBE self-referential cycle: %s\n", probeLower(t, source))
	})
}

// probeNestedTypes declares depth separately named object types, prefix0 to prefix(depth-1), each with a
// number, a function-typed property so the walk enters signatures, and next, the level below. The last
// level has no next, which ends the chain. One level is one property step on the walk's path.
func probeNestedTypes(prefix string, depth int) string {
	var types strings.Builder
	for level := range depth {
		next := ""
		if level+1 < depth {
			next = "; next: " + prefix + strconv.Itoa(level+1)
		}
		types.WriteString("type " + prefix + strconv.Itoa(level) + " = { v: number; f: () => number" + next + " };\n")
	}
	return types.String()
}

// probeLower lowers source as main.ts, as lowerSourceAs does, but answers a load error as text instead
// of failing, so the probe prints every line.
func probeLower(t *testing.T, source string) string {
	t.Helper()
	program, err := load.Load([]string{writeAgreementSource(t, "main.ts", source)})
	if err == nil {
		_, err = Lower(context.Background(), program)
	}
	if err != nil {
		return strings.ReplaceAll(err.Error(), "\n", " | ")
	}
	return "ok"
}
