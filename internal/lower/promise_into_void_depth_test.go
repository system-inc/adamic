package lower

import (
	"strconv"
	"strings"
	"testing"
)

// #96pc420: refusePromiseIntoVoid's bound is promiseIntoVoidDepth, 256 pairs along one path. The relation
// these pins drive is an async function's `return load();`, a Promise<A0> handed to a Promise<B0> return,
// where A0 and B0 are separately declared nested object types, one property step a level. It's the one
// shape this rule walks deep alone: at an initializer, an argument, an awaited return or a promise field,
// refuseVoidWidening walks the same relation first and stops at its own 64 (J-7pfy2x6). Past this rule
// the return stops at Promise adoption, which is async.go's, so no program of this shape lowers today:
// the pin past 64 holds that this rule lets it through to that next stop. Mutant: the bound put back to
// 64, and TestPromiseIntoVoidPassesPastSixtyFour turns red in the bound's words, and only it.
func promiseIntoVoidDepthSource(levels int) string {
	var source strings.Builder
	for _, prefix := range []string{"A", "B"} {
		for level := range levels {
			next := ""
			if level+1 < levels {
				next = "; next: " + prefix + strconv.Itoa(level+1)
			}
			source.WriteString("type " + prefix + strconv.Itoa(level) + " = { v: number; f: () => number" + next + " };\n")
		}
	}
	source.WriteString("function makeA(): A0 {\n\tthrow new Error('made');\n}\n")
	source.WriteString("async function load(): Promise<A0> {\n\treturn makeA();\n}\n")
	source.WriteString("async function view(): Promise<B0> {\n\treturn load();\n}\n")
	source.WriteString("const b: B0 = await view();\nconsole.log(String(b.v));\n")
	return source.String()
}

const promiseIntoVoidBoundWords = "a structural relation deeper than 256 type pairs while checking promise-returning functions seen as void-returning"

func TestPromiseIntoVoidPassesPastSixtyFour(t *testing.T) {
	t.Parallel()
	for _, extension := range []string{".a", ".ts"} {
		t.Run(extension, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSourceAs(t, promiseIntoVoidDepthSource(100), extension)
			if err == nil || strings.Contains(err.Error(), "deeper than") || !strings.Contains(err.Error(), "Promise adoption") {
				t.Fatalf("want the 100-level return through this rule to Promise adoption's stop, got %v", err)
			}
		})
	}
}

func TestPromiseIntoVoidStopsPastItsBound(t *testing.T) {
	t.Parallel()
	for _, extension := range []string{".a", ".ts"} {
		t.Run(extension, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSourceAs(t, promiseIntoVoidDepthSource(300), extension)
			if err == nil || !strings.Contains(err.Error(), promiseIntoVoidBoundWords) {
				t.Fatalf("want the 300-level return stopped in the bound's words, got %v", err)
			}
		})
	}
}

// A self-referential type ends by seen, never by the bound: the walk comes back to the pair it started
// from and answers nil. Mutant: seen dropped, and the walk runs to the bound and stops in its words.
func TestPromiseIntoVoidEndsACycleWithoutTheBound(t *testing.T) {
	t.Parallel()
	const source = "interface ListA {\n\tnext: ListA | undefined;\n\trun: () => void;\n}\ninterface ListB {\n\tnext: ListB | undefined;\n\trun: () => void;\n}\n" +
		"const a: ListA = { next: undefined, run: (): void => {} };\nconst b: ListB = a;\nb.run();\nconsole.log(String(b.next === undefined));\n"
	for _, extension := range []string{".a", ".ts"} {
		t.Run(extension, func(t *testing.T) {
			t.Parallel()
			if _, err := lowerSourceAs(t, source, extension); err != nil {
				t.Fatalf("want the self-referential pair lowered, got %v", err)
			}
		})
	}
}
